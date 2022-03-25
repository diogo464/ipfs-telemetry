package telemetry

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"git.d464.sh/adc/telemetry/telemetry/snapshot"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

func (t *TelemetryService) collectorPing() {
	h := t.host()

	for {
		if peerid, ok := collectorPingPickRandomPeer(h); !ok {
			time.Sleep(time.Second)
			continue
		} else {
			snapshot, err := collectorPingDoPing(h, peerid)
			if err == nil {
				fmt.Println(snapshot)
				t.pushSnapshot(snapshot)
			} else {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 10)
		}
	}
}

func collectorPingPickRandomPeer(h host.Host) (peer.ID, bool) {
	peers := h.Peerstore().PeersWithAddrs()
	lpeers := len(peers)
	if lpeers == 0 {
		return peer.ID(""), false
	}
	index := rand.Intn(lpeers)
	peerid := peers[index]
	return peerid, true
}

func collectorPingDoPing(h host.Host, p peer.ID) (*snapshot.Snapshot, error) {
	const PING_COUNT = 5
	const PING_TIMEOUT = 15

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*PING_TIMEOUT)
	defer cancel()

	if h.Network().Connectedness(p) != network.Connected {
		if err := h.Connect(ctx, h.Peerstore().PeerInfo(p)); err != nil {
			return nil, err
		}
	}

	durations := make([]time.Duration, 5)
	counter := 0
	cresult := ping.Ping(network.WithNoDial(ctx, "ping"), h, p)
	for result := range cresult {
		if result.Error != nil {
			return nil, result.Error
		}
		durations[counter] = result.RTT
		counter += 1
		if counter == 5 {
			break
		}
	}

	source := peer.AddrInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
	destination := h.Peerstore().PeerInfo(p)

	return snapshot.NewPingSnapshot(source, destination, durations), nil
}
