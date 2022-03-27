package snapshot

import (
	"fmt"
	"time"

	"github.com/ipfs/go-ipfs/core"
)

const TAG_ROUTING_TABLE = "routingtable"

type RoutingTable struct {
	// Number of peers in each dht bucket
	Buckets []int `json:"buckets"`
}

func (s *Snapshot) GetRoutingTable() *RoutingTable {
	rt := new(RoutingTable)
	s.decodeUnwrap(rt)
	return rt
}

func NewRoutingTable(buckets []int) *Snapshot {
	return NewSnapshot(TAG_ROUTING_TABLE, &RoutingTable{
		Buckets: buckets,
	})
}

func NewRoutingTableFromNode(n *core.IpfsNode) *Snapshot {
	rt := n.DHT.WAN.RoutingTable()
	buckets := make([]int, 0, 16)
	var index uint = 0
	for {
		n := rt.NPeersForCpl(index)
		if n == 0 {
			break
		}
		index += 1
		buckets = append(buckets, n)
	}
	return NewRoutingTable(buckets)
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
		snapshot := NewRoutingTableFromNode(c.node)
		fmt.Println("Pushing routing table snapshot")
		c.sink.Push(snapshot)
		time.Sleep(c.opts.Interval)
	}
}
