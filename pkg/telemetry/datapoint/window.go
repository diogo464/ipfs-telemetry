package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
)

var _ Datapoint = (*Window)(nil)

const WindowName = "window"

type Window struct {
	Timestamp       time.Time         `json:"timestamp"`
	WindowDuration  time.Duration     `json:"window_duration"`
	DatapointCount  map[string]uint32 `json:"datapoint.count"`
	DatapointMemory map[string]uint32 `json:"datapoint.memory"`
}

func (*Window) sealed()                   {}
func (*Window) GetName() string           { return WindowName }
func (s *Window) GetTimestamp() time.Time { return s.Timestamp }
func (s *Window) GetSizeEstimate() uint32 {
	// 18 -> 8 bytes of uuint3264 + ~10 bytes of name
	return estimateTimestampSize + uint32(len(s.DatapointCount))*18 + uint32(len(s.DatapointMemory))*18
}
func (s *Window) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Window{
			Window: WindowToPB(s),
		},
	}
}

func (s *Window) TotalCount() uint32 {
	var count uint32 = 0
	for _, v := range s.DatapointCount {
		count += v
	}
	return count
}

func (s *Window) TotalMemory() uint32 {
	var mem uint32 = 0
	for _, v := range s.DatapointMemory {
		mem += v
	}
	return mem
}

func WindowFromPB(in *pb.Window) (*Window, error) {
	return &Window{
		Timestamp:       pbutils.TimeFromPB(in.GetTimestamp()),
		WindowDuration:  pbutils.DurationFromPB(in.GetWindowDuration()),
		DatapointCount:  in.GetDatapointCount(),
		DatapointMemory: in.GetDatapointMemory(),
	}, nil
}

func WindowToPB(s *Window) *pb.Window {
	return &pb.Window{
		Timestamp:       pbutils.TimeToPB(&s.Timestamp),
		WindowDuration:  pbutils.DurationToPB(&s.WindowDuration),
		DatapointCount:  s.DatapointCount,
		DatapointMemory: s.DatapointMemory,
	}
}
