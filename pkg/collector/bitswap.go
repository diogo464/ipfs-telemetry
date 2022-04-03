package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/measurements"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-bitswap"
	bsmsg "github.com/ipfs/go-bitswap/message"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/atomic"
)

type BitswapOptions struct {
	Interval time.Duration
}

type bitswapCollector struct {
	ctx  context.Context
	opts BitswapOptions
	sink snapshot.Sink

	discovery_successes *atomic.Uint32
	discovery_failures  *atomic.Uint32
	messages_in         *atomic.Uint32
	messages_out        *atomic.Uint32
}

func RunBitswapCollector(ctx context.Context, n *core.IpfsNode, sink snapshot.Sink, opts BitswapOptions) {
	bs := n.Exchange.(*bitswap.Bitswap)
	c := &bitswapCollector{
		ctx:  ctx,
		opts: opts,
		sink: sink,

		discovery_successes: atomic.NewUint32(0),
		discovery_failures:  atomic.NewUint32(0),
		messages_in:         atomic.NewUint32(0),
		messages_out:        atomic.NewUint32(0),
	}
	bitswap.WithTracer(c)(bs)
	measurements.BitswapRegister(c)
	c.Run()
}

func (c *bitswapCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)

LOOP:
	for {
		select {
		case <-ticker.C:
			c.sink.Push(&snapshot.Bitswap{
				Timestamp:          snapshot.NewTimestamp(),
				DiscoverySucceeded: c.discovery_successes.Load(),
				DiscoveryFailed:    c.discovery_failures.Load(),
				MessagesIn:         c.messages_in.Load(),
				MessagesOut:        c.messages_out.Load(),
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

// measurements.Bitswap impl
func (c *bitswapCollector) IncDiscoverySuccess() { c.discovery_successes.Inc() }
func (c *bitswapCollector) IncDiscoveryFailure() { c.discovery_failures.Inc() }

// bitswap.Tracer impl
func (c *bitswapCollector) MessageReceived(peer.ID, bsmsg.BitSwapMessage) {
	c.messages_in.Inc()
}
func (c *bitswapCollector) MessageSent(peer.ID, bsmsg.BitSwapMessage) {
	c.messages_out.Inc()
}
