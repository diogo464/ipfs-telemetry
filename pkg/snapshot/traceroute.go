package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*TraceRoute)(nil)

const TRACE_ROUTE_NAME = "traceroute"

type TraceRoute struct {
	Timestamp time.Time
	Provider  string `json:"provider"`
	Output    []byte `json:"output"`
}

func (*TraceRoute) sealed()                   {}
func (*TraceRoute) GetName() string           { return TRACE_ROUTE_NAME }
func (t *TraceRoute) GetTimestamp() time.Time { return t.Timestamp }
func (t *TraceRoute) ToPB() *pb.Snapshot {
	return &pb.Snapshot{
		Body: &pb.Snapshot_Traceroute{
			Traceroute: TraceRouteToPB(t),
		},
	}
}

func TraceRouteFromPB(in *pb.TraceRoute) (*TraceRoute, error) {
	return &TraceRoute{
		Provider: in.GetProvider(),
		Output:   in.GetOutput(),
	}, nil
}

func TraceRouteToPB(in *TraceRoute) *pb.TraceRoute {
	// TODO: fix this
	return &pb.TraceRoute{
		Timestamp:   timestamppb.New(in.Timestamp),
		Origin:      nil,
		Destination: nil,
		Provider:    in.Provider,
		Output:      in.Output,
	}
}
