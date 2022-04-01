package collector

import (
	"context"
	"math/rand"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

type PingOptions struct {
	PingCount int
	Interval  time.Duration
	Timeout   time.Duration
}

type pingCollector struct {
	opts PingOptions
	h    host.Host
	sink snapshot.Sink
}

func RunPingCollector(h host.Host, sink snapshot.Sink, opts PingOptions) {
	c := &pingCollector{opts: opts, h: h, sink: sink}
	c.Run()
}

func (c *pingCollector) Run() {
	for {
		if peerid, ok := c.pickRandomPeer(); !ok {
			time.Sleep(time.Second)
			continue
		} else {
			if ping, err := c.ping(peerid); err == nil {
				c.sink.PushPing(ping)
			}
			time.Sleep(c.opts.Interval)
		}
	}
}

func (c *pingCollector) pickRandomPeer() (peer.ID, bool) {
	peers := c.h.Peerstore().PeersWithAddrs()
	lpeers := len(peers)
	if lpeers == 0 {
		return peer.ID(""), false
	}
	index := rand.Intn(lpeers)
	peerid := peers[index]
	return peerid, true
}

func (c *pingCollector) ping(p peer.ID) (*snapshot.Ping, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.opts.Timeout)
	defer cancel()

	if c.h.Network().Connectedness(p) != network.Connected {
		if err := c.h.Connect(ctx, c.h.Peerstore().PeerInfo(p)); err != nil {
			return nil, err
		}
	}

	durations := make([]time.Duration, c.opts.PingCount)
	counter := 0
	cresult := ping.Ping(network.WithNoDial(ctx, "ping"), c.h, p)
	for result := range cresult {
		if result.Error != nil {
			return nil, result.Error
		}
		durations[counter] = result.RTT
		counter += 1
		if counter == c.opts.PingCount {
			break
		}
	}

	source := peer.AddrInfo{
		ID:    c.h.ID(),
		Addrs: c.h.Addrs(),
	}
	destination := c.h.Peerstore().PeerInfo(p)

	return &snapshot.Ping{
		Source:      source,
		Destination: destination,
		Durations:   durations,
	}, nil
}
