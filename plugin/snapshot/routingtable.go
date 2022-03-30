package snapshot

import (
	"fmt"
	"time"

	"github.com/ipfs/go-ipfs/core"
)

type RoutingTable struct {
	// Number of peers in each dht bucket
	Buckets []uint32 `json:"buckets"`
}

func newRoutingTableFromNode(n *core.IpfsNode) *RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := make([]uint32, 0, 16)
	var index uint = 0
	for {
		n := rt.NPeersForCpl(index)
		if n == 0 {
			break
		}
		index += 1
		buckets = append(buckets, uint32(n))
	}
	return &RoutingTable{Buckets: buckets}
}

type RoutingTableOptions struct {
	Interval time.Duration
}

type RoutingTableCollector struct {
	opts RoutingTableOptions
	sink Sink
	node *core.IpfsNode
}

func NewRoutingTableCollector(n *core.IpfsNode, sink Sink, opts RoutingTableOptions) *RoutingTableCollector {
	return &RoutingTableCollector{opts: opts, sink: sink, node: n}
}

func (c *RoutingTableCollector) Run() {
	for {
		routing_table := newRoutingTableFromNode(c.node)
		fmt.Println("Pushing routing table snapshot")
		c.sink.PushRoutingTable(routing_table)
		time.Sleep(c.opts.Interval)
	}
}
