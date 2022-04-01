package collector

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-ipfs/core"
)

type RoutingTableOptions struct {
	Interval time.Duration
}

type RoutingTableCollector struct {
	opts RoutingTableOptions
	sink snapshot.Sink
	node *core.IpfsNode
}

func NewRoutingTableCollector(n *core.IpfsNode, sink snapshot.Sink, opts RoutingTableOptions) *RoutingTableCollector {
	return &RoutingTableCollector{opts: opts, sink: sink, node: n}
}

func (c *RoutingTableCollector) Run() {
	for {
		routing_table := newRoutingTableFromNode(c.node)
		c.sink.PushRoutingTable(routing_table)
		time.Sleep(c.opts.Interval)
	}
}

func newRoutingTableFromNode(n *core.IpfsNode) *snapshot.RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := rt.DumpBuckets()
	return &snapshot.RoutingTable{Buckets: buckets}
}
