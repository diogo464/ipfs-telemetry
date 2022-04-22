package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/measurements"
	"git.d464.sh/adc/telemetry/pkg/telemetry/datapoint"
	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-ipfs/core"
	"go.uber.org/atomic"
)

var _ Collector = (*bitswapCollector)(nil)
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

func NewBitswapCollector(n *core.IpfsNode) Collector {
	bs := n.Exchange.(*bitswap.Bitswap)
	return &bitswapCollector{
		bs:                  bs,
		discovery_successes: &atomic.Uint32{},
		discovery_failures:  &atomic.Uint32{},
		messages_in:         0,
		messages_out:        0,
	}
}

// Close implements Collector
func (c *bitswapCollector) Close() {
	// TODO: measurements unregister
}

// Collect implements Collector
func (c *bitswapCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	nstats := c.bs.NetworkStat()
	sink.Push(&datapoint.Bitswap{
		Timestamp:          datapoint.NewTimestamp(),
		DiscoverySucceeded: c.discovery_successes.Load(),
		DiscoveryFailed:    c.discovery_failures.Load(),
		MessagesIn:         nstats.MessagesRecvd,
		MessagesOut:        nstats.MessagesSent,
	})
}

// IncDiscoveryFailure implements measurements.Bitswap
func (c *bitswapCollector) IncDiscoveryFailure() { c.discovery_failures.Inc() }

// IncDiscoverySuccess implements measurements.Bitswap
func (c *bitswapCollector) IncDiscoverySuccess() { c.discovery_successes.Inc() }
