package datapoint

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/datapoint"
	"git.d464.sh/adc/telemetry/pkg/telemetry/pbutils"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*TraceRoute)(nil)

const TraceRouteName = "traceroute"

type TraceRoute struct {
	Timestamp   time.Time
	Origin      peer.AddrInfo `json:"origin"`
	Destination peer.AddrInfo `json:"destination"`
	Provider    string        `json:"provider"`
	Output      []byte        `json:"output"`
}

func (*TraceRoute) sealed()                   {}
func (*TraceRoute) GetName() string           { return TraceRouteName }
func (t *TraceRoute) GetTimestamp() time.Time { return t.Timestamp }
func (t *TraceRoute) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 2*estimatePeerAddrInfoSize + uint32(len(t.Provider)) + uint32(len(t.Output))
}
func (t *TraceRoute) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Traceroute{
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
