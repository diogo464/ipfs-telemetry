package snapshot

import (
	"fmt"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
)

var ErrInvalidPbType = fmt.Errorf("invalid pb type")

func NewTimestamp() time.Time {
	return time.Now().UTC()
}

//go-sumtype:decl Snapshot
type Snapshot interface {
	sealed()

	GetName() string
	GetTimestamp() time.Time
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
	case *pb.Snapshot_Bitswap:
		return BitswapFromPB(v.GetBitswap())
	case *pb.Snapshot_Storage:
		return StorageFromPB(v.GetStorage())
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
