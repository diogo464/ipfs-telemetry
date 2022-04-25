package datapoint

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/datapoint"
)

const (
	estimatePeerIdSize       = 64
	estimateMultiAddrSize    = 48
	estimateProtocolIdSize   = 32
	estimatePeerAddrInfoSize = estimatePeerIdSize + estimateMultiAddrSize*8
	estimateDurationSize     = 8
	estimateTimestampSize    = 24
	estimateMetricsStatsSize = 4 * 8
)

func NewTimestamp() time.Time {
	return time.Now().UTC()
}

//go-sumtype:decl Datapoint
type Datapoint interface {
	sealed()

	GetName() string
	GetTimestamp() time.Time
	GetSizeEstimate() uint32
	ToPB() *pb.Datapoint
}

func FromPB(v *pb.Datapoint) (Datapoint, error) {
	switch v.GetBody().(type) {
	case *pb.Datapoint_Ping:
		return PingFromPB(v.GetPing())
	case *pb.Datapoint_RoutingTable:
		return RoutingTableFromPB(v.GetRoutingTable())
	case *pb.Datapoint_Network:
		return NetworkFromPB(v.GetNetwork())
	case *pb.Datapoint_Resources:
		return ResourcesFromPB(v.GetResources())
	case *pb.Datapoint_Traceroute:
		return TraceRouteFromPB(v.GetTraceroute())
	case *pb.Datapoint_Kademlia:
		return KademliaFromPB(v.GetKademlia())
	case *pb.Datapoint_KademliaQuery:
		return KademliaQueryFromPB(v.GetKademliaQuery())
	case *pb.Datapoint_KademliaHandler:
		return KademliaHandlerFromPB(v.GetKademliaHandler())
	case *pb.Datapoint_Bitswap:
		return BitswapFromPB(v.GetBitswap())
	case *pb.Datapoint_Storage:
		return StorageFromPB(v.GetStorage())
	case *pb.Datapoint_Window:
		return WindowFromPB(v.GetWindow())
	default:
		panic("unimplemented")
	}
}

func FromArrayPB(v []*pb.Datapoint) ([]Datapoint, error) {
	out := make([]Datapoint, 0, len(v))
	for _, spb := range v {
		s, err := FromPB(spb)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}