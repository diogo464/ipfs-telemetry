package telemetry

import (
	"context"
	"runtime"

	"git.d464.sh/adc/telemetry/pkg/telemetry/pb"
	"github.com/google/uuid"
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
	if remote_sesssion == s.s {
		since = req.GetSince()
	}

	session := s.s.String()
	set := s.w.Since(since)

	return &pb.GetSnapshotsResponse{
		Session: session,
		Next:    s.w.NextSeqN(),
		Set:     set,
	}, nil
}
