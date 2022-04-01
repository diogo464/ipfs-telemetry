package collector

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-ipfs/core"
)

type RoutingTableOptions struct {
	Interval time.Duration
}

type routingTableCollector struct {
	opts RoutingTableOptions
	sink snapshot.Sink
	node *core.IpfsNode
}

func RunRoutintTableCollector(n *core.IpfsNode, sink snapshot.Sink, opts RoutingTableOptions) {
	c := &routingTableCollector{opts: opts, sink: sink, node: n}
	c.Run()
}

func (c *routingTableCollector) Run() {
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
