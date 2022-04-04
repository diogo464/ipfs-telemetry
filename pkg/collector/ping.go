package collector

import (
	"context"
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
	ctx  context.Context
	opts PingOptions
	h    host.Host
	sink snapshot.Sink
}

func RunPingCollector(ctx context.Context, h host.Host, sink snapshot.Sink, opts PingOptions) {
	c := &pingCollector{
		ctx:  ctx,
		opts: opts,
		h:    h,
		sink: sink,
	}
	c.Run()
}

func (c *pingCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)
	picker := newPeerPicker(c.h)
	defer picker.close()

LOOP:
	for {
		select {
		case <-ticker.C:
			if p, ok := picker.pick(); ok {
				if ps, err := c.ping(p); err == nil {
					c.sink.Push(ps)
				}
			}
		case <-c.ctx.Done():
			break LOOP
		}
	}
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
		Timestamp:   snapshot.NewTimestamp(),
		Source:      source,
		Destination: destination,
		Durations:   durations,
	}, nil
}
