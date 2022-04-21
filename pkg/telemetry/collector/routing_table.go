package collector

import (
	"context"

	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
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
func (c *routingTableCollector) Collect(ctx context.Context, sink snapshot.Sink) {
	routing_table := newRoutingTableFromNode(c.node)
	sink.Push(routing_table)
}

func newRoutingTableFromNode(n *core.IpfsNode) *snapshot.RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := rt.DumpBuckets()
	return &snapshot.RoutingTable{
		Timestamp: snapshot.NewTimestamp(),
		Buckets:   buckets,
	}
}
