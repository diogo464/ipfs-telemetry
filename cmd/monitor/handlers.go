package main

import (
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (s *Monitor) handleSnapshot(p peer.ID, ss snapshot.Snapshot) error {
	switch v := ss.(type) {
	case *snapshot.Ping:
		return s.handlePingSnapshot(p, v)
	case *snapshot.Network:
		return s.handleNetworkSnapshot(p, v)
	case *snapshot.RoutingTable:
		return s.handleRoutingTableSnapshot(p, v)
	default:
		return nil
	}
}

func (s *Monitor) handlePingSnapshot(p peer.ID, snapshot *snapshot.Ping) error {
	return nil
}

func (s *Monitor) handleNetworkSnapshot(p peer.ID, snapshot *snapshot.Network) error {
	return nil
}

func (s *Monitor) handleRoutingTableSnapshot(p peer.ID, snapshot *snapshot.RoutingTable) error {
	return nil
}
