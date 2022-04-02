package collector

import (
	"context"
	"fmt"
	"time"

	bt "git.d464.sh/adc/telemetry/pkg/bitswap"
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
	ctx          context.Context
	opts         BitswapOptions
	sink         snapshot.Sink
	bstelemetry  *bt.BitswapTelemetry
	messages_in  *atomic.Uint32
	messages_out *atomic.Uint32
}

func RunBitswapCollector(ctx context.Context, n *core.IpfsNode, sink snapshot.Sink, opts BitswapOptions) {
	bs := n.Exchange.(*bitswap.Bitswap)
	c := &bitswapCollector{
		ctx:          ctx,
		opts:         opts,
		sink:         sink,
		bstelemetry:  bs.Telemetry,
		messages_in:  atomic.NewUint32(0),
		messages_out: atomic.NewUint32(0),
	}
	bitswap.WithTracer(c)(bs)
	c.Run()
}

func (c *bitswapCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)

LOOP:
	for {
		select {
		case <-ticker.C:
			stats := c.bstelemetry.GetDiscoveryStats()
			c.sink.PushBitswap(&snapshot.Bitswap{
				Timestamp:          snapshot.NewTimestamp(),
				DiscoverySucceeded: uint32(stats.Succeeded),
				DiscoveryFailed:    uint32(stats.Failed),
				MessagesIn:         c.messages_in.Load(),
				MessagesOut:        c.messages_out.Load(),
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

// bitswap.Tracer impl
func (c *bitswapCollector) MessageReceived(peer.ID, bsmsg.BitSwapMessage) {
	c.messages_in.Inc()
}
func (c *bitswapCollector) MessageSent(peer.ID, bsmsg.BitSwapMessage) {
	fmt.Println("MESSAGE OUT")
	c.messages_out.Inc()
}
