package collectors

import (
	"context"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/ipfs_telemetry/pkg/traceroute"
	"github.com/diogo464/ipfs_telemetry/pkg/utils"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ telemetry.Collector = (*tracerouteCollector)(nil)

type tracerouteCollector struct {
	h      host.Host
	picker *peerPicker
}

func TraceRoute(h host.Host) telemetry.Collector {
	return &tracerouteCollector{
		h:      h,
		picker: newPeerPicker(h),
	}
}

// Descriptor implements telemetry.Collector
func (*tracerouteCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.TraceRouteName,
	}
}

// Open implements telemetry.Collector
func (*tracerouteCollector) Open() {}

// Close implements Collector
func (c *tracerouteCollector) Close() {
	c.picker.close()
}

// Collect implements Collector
func (c *tracerouteCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	if p, ok := c.picker.pick(); ok {
		s, err := c.trace(ctx, p)
		if err != nil {
			return err
		}
		return datapoint.TraceRouteSerialize(s, stream)
	}
	return nil
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
