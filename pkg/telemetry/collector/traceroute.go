package collector

import (
	"context"

	"git.d464.sh/adc/telemetry/pkg/telemetry/datapoint"
	"git.d464.sh/adc/telemetry/pkg/traceroute"
	"git.d464.sh/adc/telemetry/pkg/utils"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Collector = (*tracerouteCollector)(nil)

type tracerouteCollector struct {
	h      host.Host
	picker *peerPicker
}

func NewTracerouteCollector(h host.Host) Collector {
	return &tracerouteCollector{
		h:      h,
		picker: newPeerPicker(h),
	}
}

// Close implements Collector
func (c *tracerouteCollector) Close() {
	c.picker.close()
}

// Collect implements Collector
func (c *tracerouteCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	if p, ok := c.picker.pick(); ok {
		if s, err := c.trace(ctx, p); err == nil {
			sink.Push(s)
		}
	}
}

func (c *tracerouteCollector) trace(ctx context.Context, p peer.ID) (*datapoint.TraceRoute, error) {
	ip, err := utils.GetFirstPublicAddressFromMultiaddrs(c.h.Peerstore().Addrs(p))
	if err != nil {
		return nil, err
	}
	result, err := traceroute.Trace(ip.String())
	if err != nil {
		return nil, err
	}
	origin := peer.AddrInfo{ID: c.h.ID(), Addrs: c.h.Addrs()}
	destination := c.h.Peerstore().PeerInfo(p)

	return &datapoint.TraceRoute{
		Timestamp:   datapoint.NewTimestamp(),
		Origin:      origin,
		Destination: destination,
		Provider:    result.Provider,
		Output:      result.Output,
	}, nil
}
