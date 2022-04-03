package telemetry

import (
	"context"
	"io"
	"runtime"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/telemetry"
	"git.d464.sh/adc/telemetry/pkg/utils"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/network"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *TelemetryService) GetSystemInfo(context.Context, *emptypb.Empty) (*pb.SystemInfo, error) {
	response := &pb.SystemInfo{
		Os:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		Numcpu: uint32(runtime.NumCPU()),
	}
	return response, nil
}

func (s *TelemetryService) GetSnapshots(ctx context.Context, req *pb.GetSnapshotsRequest) (*pb.GetSnapshotsResponse, error) {
	remote_sesssion, err := uuid.Parse(req.Session)
	if err != nil {
		return nil, err
	}

	var since uint64 = 0
	if remote_sesssion == s.session {
		since = req.GetSince()
	}

	session := s.session.String()
	snapshots := s.wnd.Since(since)

	return &pb.GetSnapshotsResponse{
		Session:   session,
		Next:      s.wnd.NextSeqN(),
		Snapshots: snapshots,
	}, nil
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
