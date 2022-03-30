package wire

import (
	"runtime"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"github.com/google/uuid"
)

func NewSystemInfo(session uuid.UUID) *pb.Response {
	return &pb.Response{
		Session: session.String(),
		Body: &pb.Response_SystemInfo_{
			SystemInfo: &pb.Response_SystemInfo{
				Os:     runtime.GOOS,
				Arch:   runtime.GOARCH,
				Numcpu: uint32(runtime.NumCPU()),
			},
		},
	}
}

func NewSnapshots(session uuid.UUID, snapshots *pb.Response_Snapshots) *pb.Response {
	return &pb.Response{
		Session: session.String(),
		Body:    &pb.Response_Snapshots_{Snapshots: snapshots},
	}
}
