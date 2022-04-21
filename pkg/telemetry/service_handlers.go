package telemetry

import (
	"context"
	"io"
	"runtime"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/telemetry"
	"git.d464.sh/adc/telemetry/pkg/utils"
	"github.com/libp2p/go-libp2p-core/network"
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

func (s *TelemetryService) GetSnapshots(req *pb.GetSnapshotsRequest, stream pb.Telemetry_GetSnapshotsServer) error {
	{ //snapshots
		since := req.GetSince()
		for {
			result := s.snapshots.Fetch(since, FETCH_BLOCK_SIZE)
			since = result.FirstSeqN + uint64(len(result.Snapshots))
			if len(result.Snapshots) == 0 {
				break
			}

			err := stream.Send(&pb.GetSnapshotsResponse{
				Next:      since,
				Snapshots: result.Snapshots,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *TelemetryService) GetEvents(req *pb.GetEventsRequest, stream pb.Telemetry_GetEventsServer) error {
	{ //events
		since := req.GetSince()
		for {
			result := s.events.Fetch(since, FETCH_BLOCK_SIZE)
			since = result.FirstSeqN + uint64(len(result.Snapshots))
			if len(result.Snapshots) == 0 {
				break
			}

			err := stream.Send(&pb.GetEventsResponse{
				Next:   since,
				Events: result.Snapshots,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *TelemetryService) uploadHandler(stream network.Stream) {
	defer stream.Close()

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
