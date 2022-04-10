package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/measurements"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-ipfs/core"
	"go.uber.org/atomic"
)

type BitswapOptions struct {
	Interval time.Duration
}

type bitswapCollector struct {
	ctx  context.Context
	opts BitswapOptions
	sink snapshot.Sink
	bs   *bitswap.Bitswap

	discovery_successes *atomic.Uint32
	discovery_failures  *atomic.Uint32
	messages_in         uint64
	messages_out        uint64
}

func RunBitswapCollector(ctx context.Context, n *core.IpfsNode, sink snapshot.Sink, opts BitswapOptions) {
	bs := n.Exchange.(*bitswap.Bitswap)
	c := &bitswapCollector{
		ctx:  ctx,
		opts: opts,
		sink: sink,
		bs:   bs,

		discovery_successes: atomic.NewUint32(0),
		discovery_failures:  atomic.NewUint32(0),
		messages_in:         0,
		messages_out:        0,
	}
	measurements.BitswapRegister(c)
	c.Run()
}

func (c *bitswapCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)

LOOP:
	for {
		select {
		case <-ticker.C:
			nstats := c.bs.NetworkStat()
			c.sink.Push(&snapshot.Bitswap{
				Timestamp:          snapshot.NewTimestamp(),
				DiscoverySucceeded: c.discovery_successes.Load(),
				DiscoveryFailed:    c.discovery_failures.Load(),
				MessagesIn:         nstats.MessagesRecvd,
				MessagesOut:        nstats.MessagesSent,
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

// measurements.Bitswap impl
func (c *bitswapCollector) IncDiscoverySuccess() { c.discovery_successes.Inc() }
func (c *bitswapCollector) IncDiscoveryFailure() { c.discovery_failures.Inc() }
