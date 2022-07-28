package collectors

import (
	"context"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/ipfs/kubo/core"
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

// Descriptor implements telemetry.Collector
func (*routingTableCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.RoutingTableName,
	}
}

// Open implements telemetry.Collector
func (*routingTableCollector) Open() {
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
