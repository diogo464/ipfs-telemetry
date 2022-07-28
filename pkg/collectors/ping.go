package collectors

import (
	"context"
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
)

var _ telemetry.Collector = (*pingCollector)(nil)

type PingOptions struct {
	PingCount int
	Timeout   time.Duration
}

type pingCollector struct {
	opts   PingOptions
	h      host.Host
	picker *peerPicker
}

func Ping(h host.Host, opts PingOptions) telemetry.Collector {
	return &pingCollector{
		opts:   opts,
		h:      h,
		picker: newPeerPicker(h),
	}
}

// Descriptor implements telemetry.Collector
func (*pingCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.PingName,
	}
}

// Open implements telemetry.Collector
func (*pingCollector) Open() {
}

// Close implements Collector
func (c *pingCollector) Close() {
	c.picker.close()
}

// Collect implements Collector
func (c *pingCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	if p, ok := c.picker.pick(); ok {
		ps, err := c.ping(ctx, p)
		if err != nil {
			return err
		}
		return datapoint.PingSerialize(ps, stream)
	}
	return nil
}

func (c *pingCollector) ping(ctx context.Context, p peer.ID) (*datapoint.Ping, error) {
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

	return &datapoint.Ping{
		Timestamp:   datapoint.NewTimestamp(),
		Source:      source,
		Destination: destination,
		Durations:   durations,
	}, nil
}
