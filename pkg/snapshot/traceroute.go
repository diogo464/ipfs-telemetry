package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/pbutils"
	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*TraceRoute)(nil)

const TRACE_ROUTE_NAME = "traceroute"

type TraceRoute struct {
	Timestamp   time.Time
	Origin      peer.AddrInfo `json:"origin"`
	Destination peer.AddrInfo `json:"destination"`
	Provider    string        `json:"provider"`
	Output      []byte        `json:"output"`
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
	origin, err := pbutils.AddrInfoFromPB(in.Origin)
	if err != nil {
		return nil, err
	}
	destination, err := pbutils.AddrInfoFromPB(in.Destination)
	if err != nil {
		return nil, err
	}
	return &TraceRoute{
		Timestamp:   in.Timestamp.AsTime(),
		Origin:      origin,
		Destination: destination,
		Provider:    in.GetProvider(),
		Output:      in.GetOutput(),
	}, nil
}

func TraceRouteToPB(in *TraceRoute) *pb.TraceRoute {
	return &pb.TraceRoute{
		Timestamp:   timestamppb.New(in.Timestamp),
		Origin:      pbutils.AddrInfoToPB(&in.Origin),
		Destination: pbutils.AddrInfoToPB(&in.Destination),
		Provider:    in.Provider,
		Output:      in.Output,
	}
}
