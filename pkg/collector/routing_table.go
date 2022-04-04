package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-ipfs/core"
)

type RoutingTableOptions struct {
	Interval time.Duration
}

type routingTableCollector struct {
	ctx  context.Context
	opts RoutingTableOptions
	sink snapshot.Sink
	node *core.IpfsNode
}

func RunRoutingTableCollector(ctx context.Context, n *core.IpfsNode, sink snapshot.Sink, opts RoutingTableOptions) {
	c := &routingTableCollector{ctx: ctx, opts: opts, sink: sink, node: n}
	c.Run()
}

func (c *routingTableCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)
LOOP:
	for {
		select {
		case <-ticker.C:
			routing_table := newRoutingTableFromNode(c.node)
			c.sink.Push(routing_table)
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

func newRoutingTableFromNode(n *core.IpfsNode) *snapshot.RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := rt.DumpBuckets()
	return &snapshot.RoutingTable{
		Timestamp: snapshot.NewTimestamp(),
		Buckets:   buckets,
	}
}
