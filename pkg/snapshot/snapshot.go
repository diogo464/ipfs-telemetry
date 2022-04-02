package snapshot

import (
	"fmt"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
)

var ErrInvalidPbType = fmt.Errorf("invalid pb type")

//go-sumtype:decl Snapshot
type Snapshot interface {
	sealed()
}

func FromPB(v interface{}) (Snapshot, error) {
	// TODO: make sure this is up to date
	switch s := v.(type) {
	case *pb.Ping:
		return PingFromPB(s)
	case *pb.RoutingTable:
		return RoutingTableFromPB(s)
	case *pb.Network:
		return NetworkFromPB(s)
	case *pb.Resources:
		return ResourcesFromPB(s)
	case *pb.TraceRoute:
		panic("unimplemented")
	case *pb.Bitswap:
		return BitswapFromPB(s)
	default:
		return nil, ErrInvalidPbType
	}
}

func SetPBToSnapshotArray(in *pb.Set) ([]Snapshot, error) {
	capacity := len(in.GetPings()) + len(in.GetRoutingTables()) + len(in.GetNetworks()) + len(in.GetResources()) + len(in.GetTraceroutes()) + len(in.GetBitswaps())
	snapshots := make([]Snapshot, 0, capacity)

	for _, v := range in.GetPings() {
		ss, err := FromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ss)
	}
	for _, v := range in.GetRoutingTables() {
		ss, err := FromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ss)
	}
	for _, v := range in.GetNetworks() {
		ss, err := FromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ss)
	}
	for _, v := range in.GetResources() {
		ss, err := FromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ss)
	}
	for _, v := range in.GetTraceroutes() {
		ss, err := FromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ss)
	}
	for _, v := range in.GetBitswaps() {
		ss, err := FromPB(v)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, ss)
	}

	return snapshots, nil
}

func NewTimestamp() time.Time {
	return time.Now().UTC()
}

type Set struct {
	Ping         []*Ping
	Network      []*Network
	RoutingTable []*RoutingTable
}

func NewSet() *Set {
	return &Set{}
}

func (s *Set) PushPing(p *Ping)                 { s.Ping = append(s.Ping, p) }
func (s *Set) PushRoutingTable(r *RoutingTable) { s.RoutingTable = append(s.RoutingTable, r) }
func (s *Set) PushNetwork(n *Network)           { s.Network = append(s.Network, n) }

func (s *Set) ToPB() *pb.Set {
	return &pb.Set{
		Pings:         PingArrayToPB(s.Ping),
		RoutingTables: RoutingTableArrayToPB(s.RoutingTable),
		Networks:      NetworkArrayToPB(s.Network),
	}
}
