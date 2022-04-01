package main

import (
	"encoding/json"
	"fmt"

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
	case *snapshot.Resources:
		return s.handleResourcesSnapshot(p, v)
	case *snapshot.Bitswap:
		return s.handleBitswapSnapshot(p, v)
	default:
		panic("unimplemented")
	}
}

func (s *Monitor) handlePingSnapshot(p peer.ID, ss *snapshot.Ping) error {
	return nil
}

func (s *Monitor) handleNetworkSnapshot(p peer.ID, ss *snapshot.Network) error {
	return nil
}

func (s *Monitor) handleRoutingTableSnapshot(p peer.ID, ss *snapshot.RoutingTable) error {
	return nil
}

func (s *Monitor) handleResourcesSnapshot(p peer.ID, ss *snapshot.Resources) error {
	return nil
}

func (s *Monitor) handleBitswapSnapshot(p peer.ID, ss *snapshot.Bitswap) error {
	marshaled, err := json.MarshalIndent(ss, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(marshaled))
	return nil
}
