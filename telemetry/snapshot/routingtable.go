package snapshot

import "github.com/ipfs/go-ipfs/core"

type RoutingTableSnapshot struct {
	// Number of peers in each dht bucket
	Buckets []int `json:"buckets"`
}

func NewRoutingTableSnapshot(buckets []int) *Snapshot {
	return NewSnapshot("routingtable", &RoutingTableSnapshot{
		Buckets: buckets,
	})
}

func NewRoutingTableSnapshotFromNode(n *core.IpfsNode) *Snapshot {
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
	return NewRoutingTableSnapshot(buckets)
}
