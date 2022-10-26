package telemetry

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/diogo464/telemetry/pb"
	"github.com/diogo464/telemetry/utils"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc/codes"
	grpc_peer "google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const (
	TAG_UPLOAD     = "upload"
	TAG_DOWNLOAD   = "download"
	TAG_STREAM     = "getdatapoints"
	TAG_GETRECORDS = "getrecords"
)

const (
	observerDefaultCapacity = 16
	maxGrpcMessageSize      = 2 * 1024 * 1024
)

var (
	ErrBlocked              = fmt.Errorf("blocked")
	ErrStreamNotAvailable   = fmt.Errorf("stream not available")
	ErrPropertyNotAvailable = fmt.Errorf("property not available")
	ErrNotEnabled           = fmt.Errorf("not enabled")
)

var _ io.Writer = (*propertyWriter)(nil)

type propertyWriter struct {
	srv pb.Telemetry_GetPropertyServer
}

// Write implements io.Writer
func (w propertyWriter) Write(p []byte) (n int, err error) {
	written := 0
	for len(p) > 0 {
		chunkLen := utils.Min(len(p), maxGrpcMessageSize)
		err := w.srv.Send(&pb.PropertySegment{
			Data: p[:chunkLen],
		})
		if err != nil {
			return written, err
		}
		p = p[chunkLen:]
		written += chunkLen
	}
	return written, nil
}

func (s *Service) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	return &pb.GetSessionResponse{
		Uuid: s.session.String(),
	}, nil
}

func (s *Service) GetMetrics(req *pb.GetMetricsRequest, srv pb.Telemetry_GetMetricsServer) error {
	stream := s.metrics.stream
	return grpcSendStreamSegments(stream, int(req.GetSince()), srv)
}

func (s *Service) GetAvailableEvents(req *pb.GetAvailableEventsRequest, srv pb.Telemetry_GetAvailableEventsServer) error {
	for name := range s.events.events {
		err := srv.Send(&pb.AvailableEvent{
			Name: name,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetEvent(req *pb.GetEventRequest, srv pb.Telemetry_GetEventServer) error {
	name := req.GetName()
	since := int(req.GetSeqn())
	event := s.events.events[name]
	if event == nil {
		return status.Errorf(codes.NotFound, "event does not exist")
	}
	stream := event.stream
	return grpcSendStreamSegments(stream, since, srv)
}

func (s *Service) GetAvailableSnapshots(req *pb.GetAvailableSnapshotsRequest, srv pb.Telemetry_GetAvailableSnapshotsServer) error {
	for name := range s.snapshots.snapshots {
		err := srv.Send(&pb.AvailableSnapshot{
			Name: name,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetSnapshot(req *pb.GetSnapshotRequest, srv pb.Telemetry_GetSnapshotServer) error {
	name := req.GetName()
	since := int(req.GetSeqn())
	snapshot := s.snapshots.snapshots[name]
	if snapshot == nil {
		return status.Errorf(codes.NotFound, "snapshot not found")
	}
	stream := snapshot.stream
	return grpcSendStreamSegments(stream, since, srv)
}

func (s *Service) GetAvailableProperties(req *pb.GetAvailablePropertiesRequest, srv pb.Telemetry_GetAvailablePropertiesServer) error {
	properties := s.properties
	for _, entry := range properties.properties {
		err := srv.Send(&pb.AvailableProperty{
			Name:     entry.descriptor.Name,
			Encoding: uint32(entry.descriptor.Encoding),
			Constant: entry.descriptor.Constant,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetProperty(req *pb.GetPropertyRequest, srv pb.Telemetry_GetPropertyServer) error {
	properties := s.properties
	propertyEntry := properties.properties[req.GetProperty()]
	if propertyEntry == nil {
		return ErrPropertyNotAvailable
	}

	writer := propertyWriter{srv: srv}
	return propertyEntry.collector.Collect(srv.Context(), writer)
}

//func (s *Service) GetDebug(ctx context.Context, req *types.Empty) (*pb.TelemetryDebug, error) {
//	if !s.opts.enableDebug {
//		return nil, ErrNotEnabled
//	}
//
//	//for streamName, streamEntry := range s.streams {
//	//	stream := streamEntry.stream
//	//	fmt.Println(streamName)
//	//	fmt.Println("\tDefault buffer size: ", stream.opts.defaultBufferSize)
//	//	fmt.Println("\tActive buffer lifetime: ", stream.opts.activeBufferLifetime)
//	//	fmt.Println("\tSegments: ", stream.segments.Len())
//	//	fmt.Println("\tBuffer pool len: ", stream.bufferPool.len())
//	//}
//
//	streamDebugs := make([]*pb.TelemetryDebug_Stream, 0, len(s.streams))
//	for streamName, streamEntry := range s.streams {
//		stream := streamEntry.stream
//		dbg := stream.debug()
//		streamDebugs = append(streamDebugs, &pb.TelemetryDebug_Stream{
//			Name:  streamName,
//			Used:  dbg.usedSize,
//			Total: dbg.totalSize,
//		})
//	}
//
//	now := time.Now().UTC()
//	return &pb.TelemetryDebug{
//		Timestamp: utils.TimeToPB(&now),
//		Streams:   streamDebugs,
//	}, nil
//}

func (s *Service) uploadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.uploadBlocker.isBlocked(publicIp) {
			_ = utils.WriteU32(stream, 0)
			return
		}
		s.uploadBlocker.block(publicIp, BLOCK_DURATION_BANDWIDTH)
	}

	requested_payload, err := utils.ReadU32(stream)
	if err != nil || requested_payload > MAX_PAYLOAD_SIZE {
		return
	}

	upload_start := time.Now()
	n, err := io.Copy(stream, io.LimitReader(utils.NullReader{}, int64(requested_payload)))
	if err != nil {
		return
	}
	elapsed := time.Since(upload_start)
	if err != nil {
		return
	}
	rate := uint32(float64(n) / elapsed.Seconds())
	_ = utils.WriteU32(stream, rate)
}

func (s *Service) downloadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.downloadBlocker.isBlocked(publicIp) {
			_ = utils.WriteU32(stream, 0)
			return
		}
		s.downloadBlocker.block(publicIp, BLOCK_DURATION_BANDWIDTH)
	}

	expected_payload, err := utils.ReadU32(stream)
	if err != nil || expected_payload > MAX_PAYLOAD_SIZE {
		return
	}

	download_start := time.Now()
	n, err := io.Copy(io.Discard, io.LimitReader(stream, int64(expected_payload)))
	elapsed := time.Since(download_start)
	if err != nil {
		return
	}
	rate := uint32(float64(n) / elapsed.Seconds())
	_ = utils.WriteU32(stream, rate)
}

func getPublicIpFromContext(h host.Host, ctx context.Context) (net.IP, error) {
	grpcPeer, ok := grpc_peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to obtain peer")
	}
	// https://github.com/libp2p/go-libp2p-gostream/blob/master/addr.go
	pidB58 := grpcPeer.Addr.String()
	pid, err := peer.Decode(pidB58)
	if err != nil {
		return nil, err
	}
	addrs := h.Peerstore().Addrs(pid)
	return utils.GetFirstPublicAddressFromMultiaddrs(addrs)
}

type grpcSegmentSender interface {
	Send(*pb.StreamSegment) error
}

func grpcSendStreamSegments(stream *Stream, since int, srv grpcSegmentSender) error {
	for {
		segments := stream.Segments(since, 128)
		if len(segments) == 0 {
			break
		}
		since += len(segments)
		for _, segment := range segments {
			err := srv.Send(&pb.StreamSegment{
				Seqn: uint32(segment.SeqN),
				Data: segment.Data,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
