package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*Window)(nil)

const WindowName = "window"

type Window struct {
	Timestamp      time.Time         `json:"timestamp"`
	WindowDuration time.Duration     `json:"window_duration"`
	SnapshotCount  map[string]uint32 `json:"snapshot_count"`
	SnapshotMemory map[string]uint32 `json:"snapshot_memory"`
}

func (*Window) sealed()                   {}
func (*Window) GetName() string           { return WindowName }
func (s *Window) GetTimestamp() time.Time { return s.Timestamp }
func (s *Window) GetSizeEstimate() uint32 {
	// 18 -> 8 bytes of uuint3264 + ~10 bytes of name
	return estimateTimestampSize + uint32(len(s.SnapshotCount))*18 + uint32(len(s.SnapshotMemory))*18
}
func (s *Window) ToPB() *pb.Snapshot {
	return &pb.Snapshot{
		Body: &pb.Snapshot_Window{
			Window: WindowToPB(s),
		},
	}
}

func (s *Window) TotalCount() uint32 {
	var count uint32 = 0
	for _, v := range s.SnapshotCount {
		count += v
	}
	return count
}

func (s *Window) TotalMemory() uint32 {
	var mem uint32 = 0
	for _, v := range s.SnapshotMemory {
		mem += v
	}
	return mem
}

func WindowFromPB(in *pb.Window) (*Window, error) {
	return &Window{
		Timestamp:      in.GetTimestamp().AsTime(),
		WindowDuration: in.GetWindowDuration().AsDuration(),
		SnapshotCount:  in.GetSnapshotCount(),
		SnapshotMemory: in.GetSnapshotMemory(),
	}, nil
}

func WindowToPB(s *Window) *pb.Window {
	return &pb.Window{
		Timestamp:      timestamppb.New(s.Timestamp),
		WindowDuration: durationpb.New(s.WindowDuration),
		SnapshotCount:  s.SnapshotCount,
		SnapshotMemory: s.SnapshotMemory,
	}
}
