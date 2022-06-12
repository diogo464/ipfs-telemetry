package collectors

import (
	"context"

	"github.com/diogo464/telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/ipfs/go-ipfs/core"
)

var _ telemetry.Collector = (*routingTableCollector)(nil)

type routingTableCollector struct {
	node *core.IpfsNode
}

func RoutingTable(n *core.IpfsNode) telemetry.Collector {
	return &routingTableCollector{
		node: n,
	}
}

// Name implements telemetry.Collector
func (*routingTableCollector) Name() string {
	return "Routing Table"
}

// Close implements Collector
func (*routingTableCollector) Close() {
}

// Collect implements Collector
func (c *routingTableCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	routing_table := newRoutingTableFromNode(c.node)
	return datapoint.RoutingTableSerialize(routing_table, stream)
}

func newRoutingTableFromNode(n *core.IpfsNode) *datapoint.RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := rt.DumpBuckets()
	return &datapoint.RoutingTable{
		Timestamp: datapoint.NewTimestamp(),
		Buckets:   buckets,
	}
}
