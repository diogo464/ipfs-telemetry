package collector

import (
	"context"
	"time"

	bt "git.d464.sh/adc/telemetry/pkg/bitswap"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-bitswap"
	"github.com/ipfs/go-ipfs/core"
)

type BitswapOptions struct {
	Interval time.Duration
}

type bitswapCollector struct {
	ctx         context.Context
	opts        BitswapOptions
	sink        snapshot.Sink
	bstelemetry *bt.BitswapTelemetry
}

func RunBitswapCollector(ctx context.Context, n *core.IpfsNode, sink snapshot.Sink, opts BitswapOptions) {
	bs := n.Exchange.(*bitswap.Bitswap)
	c := &bitswapCollector{
		ctx:         ctx,
		opts:        opts,
		sink:        sink,
		bstelemetry: bs.Telemetry,
	}
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
				Timestamp:          time.Now(),
				DiscoverySucceeded: uint32(stats.Succeeded),
				DiscoveryFailed:    uint32(stats.Failed),
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}
