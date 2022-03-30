package main

import (
	"fmt"

	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (s *Monitor) handleSnapshot(p peer.ID, snapshot telemetry.Snapshot) error {
	switch v := snapshot.(type) {
	case *telemetry.PingSnapshot:
		return s.handlePingSnapshot(p, v)
	case *telemetry.NetworkSnapshot:
		return s.handleNetworkSnapshot(p, v)
	case *telemetry.RoutingTableSnapshot:
		return s.handleRoutingTableSnapshot(p, v)
	default:
		return nil
	}
}

func (s *Monitor) handlePingSnapshot(p peer.ID, snapshot *telemetry.PingSnapshot) error {
	fmt.Println("Handling PingSnapshot")
	return nil
}

func (s *Monitor) handleNetworkSnapshot(p peer.ID, snapshot *telemetry.NetworkSnapshot) error {
	fmt.Println("Handling PingSnapshot")
	return nil
}

func (s *Monitor) handleRoutingTableSnapshot(p peer.ID, snapshot *telemetry.RoutingTableSnapshot) error {
	fmt.Println("Handling PingSnapshot")
	return nil
}
