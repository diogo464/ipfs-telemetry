package telemetry

import (
	"context"
	"io"
	"runtime"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/telemetry"
	"github.com/diogo464/telemetry/pkg/utils"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *TelemetryService) GetSessionInfo(context.Context, *emptypb.Empty) (*pb.GetSessionInfoResponse, error) {
	response := &pb.GetSessionInfoResponse{
		Session:  s.session.String(),
		BootTime: timestamppb.New(s.boot_time),
	}
	return response, nil
}

func (s *TelemetryService) GetSystemInfo(context.Context, *emptypb.Empty) (*pb.SystemInfo, error) {
	response := &pb.SystemInfo{
		Os:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Numcpu: uint32(runtime.NumCPU()),
	}
	return response, nil
}

func (s *TelemetryService) GetDatapoints(req *pb.GetDatapointsRequest, stream pb.Telemetry_GetDatapointsServer) error {
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

func (s *TelemetryService) uploadHandler(stream network.Stream) {
	defer stream.Close()

	if publicIp, err := utils.GetFirstPublicAddressFromMultiaddrs([]multiaddr.Multiaddr{stream.Conn().RemoteMultiaddr()}); err == nil {
		if s.throttler_upload.isAllowed(publicIp) {
			s.throttler_upload.disallow(publicIp, BANDWIDTH_BLOCK_DURATION)
		} else {
			utils.WriteU32(stream, 0)
			return
		}
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
		if s.throttler_download.isAllowed(publicIp) {
			s.throttler_download.disallow(publicIp, BANDWIDTH_BLOCK_DURATION)
		} else {
			utils.WriteU32(stream, 0)
			return
		}
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
