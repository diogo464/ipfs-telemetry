package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var _ Collector = (*pingCollector)(nil)

type PingOptions struct {
	PingCount int
	Timeout   time.Duration
}

type pingCollector struct {
	opts   PingOptions
	h      host.Host
	picker *peerPicker
}

func NewPingCollector(h host.Host, opts PingOptions) Collector {
	return &pingCollector{
		opts:   opts,
		h:      h,
		picker: newPeerPicker(h),
	}
}

// Close implements Collector
func (*pingCollector) Close() {
}

// Collect implements Collector
func (c *pingCollector) Collect(ctx context.Context, sink snapshot.Sink) {
	if p, ok := c.picker.pick(); ok {
		if ps, err := c.ping(ctx, p); err == nil {
			sink.Push(ps)
		}
	}
}

func (c *pingCollector) ping(ctx context.Context, p peer.ID) (*snapshot.Ping, error) {
	ctx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
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
