package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
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

//go-sumtype:decl Snapshot
type Snapshot interface {
	sealed()

	GetName() string
	GetTimestamp() time.Time
	GetSizeEstimate() uint32
	ToPB() *pb.Snapshot
}

func FromPB(v *pb.Snapshot) (Snapshot, error) {
	switch v.GetBody().(type) {
	case *pb.Snapshot_Ping:
		return PingFromPB(v.GetPing())
	case *pb.Snapshot_RoutingTable:
		return RoutingTableFromPB(v.GetRoutingTable())
	case *pb.Snapshot_Network:
		return NetworkFromPB(v.GetNetwork())
	case *pb.Snapshot_Resources:
		return ResourcesFromPB(v.GetResources())
	case *pb.Snapshot_Traceroute:
		return TraceRouteFromPB(v.GetTraceroute())
	case *pb.Snapshot_Kademlia:
		return KademliaFromPB(v.GetKademlia())
	case *pb.Snapshot_KademliaQuery:
		return KademliaQueryFromPB(v.GetKademliaQuery())
	case *pb.Snapshot_KademliaHandler:
		return KademliaHandlerFromPB(v.GetKademliaHandler())
	case *pb.Snapshot_Bitswap:
		return BitswapFromPB(v.GetBitswap())
	case *pb.Snapshot_Storage:
		return StorageFromPB(v.GetStorage())
	case *pb.Snapshot_Window:
		return WindowFromPB(v.GetWindow())
	default:
		panic("unimplemented")
	}
}

func FromArrayPB(v []*pb.Snapshot) ([]Snapshot, error) {
	out := make([]Snapshot, 0, len(v))
	for _, spb := range v {
		s, err := FromPB(spb)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
