package telemetry

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"runtime"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
	"github.com/diogo464/telemetry/pkg/utils"
	"github.com/gogo/protobuf/types"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	grpc_peer "google.golang.org/grpc/peer"
)

const (
	TAG_UPLOAD     = "upload"
	TAG_DOWNLOAD   = "download"
	TAG_STREAM     = "getdatapoints"
	TAG_GETRECORDS = "getrecords"
)

var (
	ErrBlocked            = fmt.Errorf("blocked")
	ErrStreamNotAvailable = fmt.Errorf("stream not available")
	ErrNotEnabled         = fmt.Errorf("not enabled")
)

func (s *Service) GetSessionInfo(context.Context, *types.Empty) (*pb.GetSessionInfoResponse, error) {
	response := &pb.GetSessionInfoResponse{
		Session:  s.session.String(),
		BootTime: pbutils.TimeToPB(&s.bootTime),
	}
	return response, nil
}

func (s *Service) GetSystemInfo(context.Context, *types.Empty) (*pb.SystemInfo, error) {
	response := &pb.SystemInfo{
		Os:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Numcpu: uint32(runtime.NumCPU()),
	}
	return response, nil
}

func (s *Service) GetAvailableStreams(ctx context.Context, req *pb.GetAvailableStreamsRequest) (*pb.GetAvailableStreamsResponse, error) {
	streams := make([]string, 0, len(s.streams))
	for k := range s.streams {
		streams = append(streams, k)
	}
	return &pb.GetAvailableStreamsResponse{
		Streams: streams,
	}, nil
}

func (s *Service) GetStream(req *pb.GetStreamRequest, srv pb.Telemetry_GetStreamServer) error {
	streamEntry := s.streams[req.GetStream()]
	if streamEntry == nil {
		return ErrStreamNotAvailable
	}

	if publicIp, err := getPublicIpFromContext(s.node.PeerHost, srv.Context()); err == nil {
		if streamEntry.blocker.isBlocked(publicIp) {
			return ErrBlocked
		}
		streamEntry.blocker.block(publicIp, BLOCK_DURATION_STREAM)
	}
	stream := streamEntry.stream

	segments := stream.Segments(int(req.GetSeqn()), math.MaxInt)
	for _, segment := range segments {
		err := srv.Send(&pb.StreamSegment{
			Seqn: uint32(segment.SeqN),
			Data: segment.Data,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetProviderRecords(_ *pb.GetProviderRecordsRequest, stream pb.Telemetry_GetProviderRecordsServer) error {
	if publicIp, err := getPublicIpFromContext(s.node.PeerHost, stream.Context()); err == nil {
		if s.provRecordsBlocker.isBlocked(publicIp) {
			return ErrBlocked
		}
		s.provRecordsBlocker.block(publicIp, BLOCK_DURATION_GETRECORDPROVIDERS)
	}

	records, err := s.node.DHT.WAN.ProviderStore().GetProviderRecords(stream.Context())
	if err != nil {
		return err
	}

	for record := range records {
		pid, err := record.Peer.MarshalBinary()
		if err != nil {
			return err
		}

		err = stream.Send(&pb.ProviderRecord{
			Key:         record.Key,
			Peer:        pid,
			LastRefresh: pbutils.TimeToPB(&record.LastRefresh),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetDebug(ctx context.Context, req *types.Empty) (*pb.TelemetryDebug, error) {
	if !s.opts.enableDebug {
		return nil, ErrNotEnabled
	}

	for streamName, streamEntry := range s.streams {
		stream := streamEntry.stream
		fmt.Println(streamName)
		fmt.Println("\tDefault buffer size: ", stream.opts.defaultBufferSize)
		fmt.Println("\tActive buffer lifetime: ", stream.opts.activeBufferLifetime)
		fmt.Println("\tSegments: ", stream.segments.Len())
		fmt.Println("\tBuffer pool len: ", stream.bufferPool.len())
	}

	streamDebugs := make([]*pb.TelemetryDebug_Stream, 0, len(s.streams))
	for streamName, streamEntry := range s.streams {
		stream := streamEntry.stream
		dbg := stream.debug()
		streamDebugs = append(streamDebugs, &pb.TelemetryDebug_Stream{
			Name:  streamName,
			Used:  dbg.usedSize,
			Total: dbg.totalSize,
		})
	}

	now := time.Now().UTC()
	return &pb.TelemetryDebug{
		Timestamp: pbutils.TimeToPB(&now),
		Streams:   streamDebugs,
	}, nil
}

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
