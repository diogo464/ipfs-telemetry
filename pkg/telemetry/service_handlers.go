package telemetry

import (
	"context"
	"fmt"
	"io"
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
	TAG_DATAPOINTS = "getdatapoints"
	TAG_GETRECORDS = "getrecords"
)

var errBlocked error = fmt.Errorf("blocked")

func (s *TelemetryService) GetSessionInfo(context.Context, *types.Empty) (*pb.GetSessionInfoResponse, error) {
	response := &pb.GetSessionInfoResponse{
		Session:  s.session.String(),
		BootTime: pbutils.TimeToPB(&s.boot_time),
	}
	return response, nil
}

func (s *TelemetryService) GetSystemInfo(context.Context, *types.Empty) (*pb.SystemInfo, error) {
	response := &pb.SystemInfo{
		Os:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Numcpu: uint32(runtime.NumCPU()),
	}
	return response, nil
}

func (s *TelemetryService) GetDatapoints(req *pb.GetDatapointsRequest, stream pb.Telemetry_GetDatapointsServer) error {
	if publicIp, err := getPublicIpFromContext(s.node.PeerHost, stream.Context()); err == nil {
		if s.requestBlocker.isBlocked(TAG_DATAPOINTS, publicIp) {
			return errBlocked
		}
		s.requestBlocker.block(TAG_DATAPOINTS, publicIp, BLOCK_DURATION_GETDATAPOINTS)
		defer s.requestBlocker.block(TAG_DATAPOINTS, publicIp, BLOCK_DURATION_GETDATAPOINTS)
	}

	since := req.GetSince()
	sleep := time.Duration((1.0 / (float64(DATAPOINT_UPLOAD_RATE) / float64(DATAPOINT_FETCH_BLOCK_SIZE))) * float64(time.Second))
	for {
		time.Sleep(sleep)
		result := s.twindow.Fetch(since, DATAPOINT_FETCH_BLOCK_SIZE)
		if len(result.Datapoints) == 0 {
			break
		}
		since = result.NextSeqN

		err := stream.Send(&pb.GetDatapointsResponse{
			Next:       result.NextSeqN,
			Datapoints: result.Datapoints,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TelemetryService) GetProviderRecords(_ *types.Empty, stream pb.Telemetry_GetProviderRecordsServer) error {
	records, err := s.node.DHT.WAN.ProviderStore().GetProviderRecords(stream.Context())
	if err != nil {
		return err
	}

	if publicIp, err := getPublicIpFromContext(s.node.PeerHost, stream.Context()); err == nil {
		if s.requestBlocker.isBlocked(TAG_GETRECORDS, publicIp) {
			return errBlocked
		}
		s.requestBlocker.block(TAG_GETRECORDS, publicIp, BLOCK_DURATION_GETRECORDPROVIDERS)
		defer s.requestBlocker.block(TAG_GETRECORDS, publicIp, BLOCK_DURATION_GETRECORDPROVIDERS)
	}

	for _, record := range records {
		pbentries := make([]*pb.ProviderRecord_Entry, len(record.Entries))
		for i, entry := range record.Entries {
			pbentries[i] = &pb.ProviderRecord_Entry{
				Peer:        entry.Peer.String(),
				LastRefresh: pbutils.TimeToPB(&entry.LastRefresh),
			}
		}

		err = stream.Send(&pb.ProviderRecord{
			Key:     record.Key,
			Entries: pbentries,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *TelemetryService) uploadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.requestBlocker.isBlocked(TAG_UPLOAD, publicIp) {
			_ = utils.WriteU32(stream, 0)
			return
		}
		s.requestBlocker.block(TAG_UPLOAD, publicIp, BLOCK_DURATION_BANDWIDTH)
		defer s.requestBlocker.block(TAG_UPLOAD, publicIp, BLOCK_DURATION_BANDWIDTH)
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

func (s *TelemetryService) downloadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.requestBlocker.isBlocked(TAG_DOWNLOAD, publicIp) {
			_ = utils.WriteU32(stream, 0)
			return
		}
		s.requestBlocker.block(TAG_DOWNLOAD, publicIp, BLOCK_DURATION_BANDWIDTH)
		defer s.requestBlocker.block(TAG_DOWNLOAD, publicIp, BLOCK_DURATION_BANDWIDTH)
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
