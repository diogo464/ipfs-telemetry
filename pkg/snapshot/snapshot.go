package snapshot

import pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"

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

func (s *Set) ToPB() *pb.Set {
	return &pb.Set{
		Pings:         PingArrayToPB(s.Ping),
		RoutingTables: RoutingTableArrayToPB(s.RoutingTable),
		Networks:      NetworkArrayToPB(s.Network),
	}
}
