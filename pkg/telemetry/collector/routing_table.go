package collector

import (
	"context"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/ipfs/go-ipfs/core"
)

var _ Collector = (*routingTableCollector)(nil)

type routingTableCollector struct {
	node *core.IpfsNode
}

func NewRoutingTableCollector(n *core.IpfsNode) Collector {
	return &routingTableCollector{
		node: n,
	}
}

// Close implements Collector
func (*routingTableCollector) Close() {
}

// Collect implements Collector
func (c *routingTableCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	routing_table := newRoutingTableFromNode(c.node)
	sink.Push(routing_table)
}

func newRoutingTableFromNode(n *core.IpfsNode) *datapoint.RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := rt.DumpBuckets()
	return &datapoint.RoutingTable{
		Timestamp: datapoint.NewTimestamp(),
		Buckets:   buckets,
	}
}
