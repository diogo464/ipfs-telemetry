package snapshot

import "git.d464.sh/adc/telemetry/pkg/telemetry/pb"

type Snapshot interface{}

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

func (s *Set) ToPB() *pb.Snapshot_Set {
	return &pb.Snapshot_Set{
		Pings:         PingArrayToPB(s.Ping),
		RoutingTables: RoutingTableArrayToPB(s.RoutingTable),
		Networks:      NetworkArrayToPB(s.Network),
	}
}
