package collectors

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/measurements"
	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-ipfs/core"
	"go.uber.org/atomic"
)

var _ telemetry.Collector = (*bitswapCollector)(nil)
var _ measurements.Bitswap = (*bitswapCollector)(nil)

type BitswapOptions struct {
	Interval time.Duration
}

type bitswapCollector struct {
	bs *bitswap.Bitswap

	discovery_successes *atomic.Uint32
	discovery_failures  *atomic.Uint32
	messages_in         uint64
	messages_out        uint64
}

func Bitswap(n *core.IpfsNode) telemetry.Collector {
	bs := n.Exchange.(*bitswap.Bitswap)
	c := &bitswapCollector{
		bs:                  bs,
		discovery_successes: &atomic.Uint32{},
		discovery_failures:  &atomic.Uint32{},
		messages_in:         0,
		messages_out:        0,
	}
	measurements.BitswapRegister(c)
	return c
}

// Name implements telemetry.Collector
func (*bitswapCollector) Name() string {
	return "Bitswap"
}

// Close implements Collector
func (c *bitswapCollector) Close() {
	// TODO: measurements unregister
}

// Collect implements Collector
func (c *bitswapCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	nstats := c.bs.NetworkStat()
	dp := &datapoint.Bitswap{
		Timestamp:          datapoint.NewTimestamp(),
		DiscoverySucceeded: c.discovery_successes.Load(),
		DiscoveryFailed:    c.discovery_failures.Load(),
		MessagesIn:         nstats.MessagesRecvd,
		MessagesOut:        nstats.MessagesSent,
	}
	return datapoint.BitswapSerialize(dp, stream)
}

// IncDiscoveryFailure implements measurements.Bitswap
func (c *bitswapCollector) IncDiscoveryFailure() { c.discovery_failures.Inc() }

// IncDiscoverySuccess implements measurements.Bitswap
func (c *bitswapCollector) IncDiscoverySuccess() { c.discovery_successes.Inc() }