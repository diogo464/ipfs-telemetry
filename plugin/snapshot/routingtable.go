package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"github.com/ipfs/go-ipfs/core"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RoutingTable struct {
	Timestamp time.Time   `json:"timestamp"`
	Buckets   [][]peer.ID `json:"buckets"`
}

func (r *RoutingTable) ToPB() *pb.Snapshot_RoutingTable {
	buckets := make([]*pb.Snapshot_RoutingTable_Bucket, 0, len(r.Buckets))
	for _, b := range r.Buckets {
		peers := make([]string, 0, len(b))
		for _, p := range b {
			peers = append(peers, p.Pretty())
		}
		buckets = append(buckets, &pb.Snapshot_RoutingTable_Bucket{
			Peers: peers,
		})
	}
	return &pb.Snapshot_RoutingTable{
		Timestamp: timestamppb.New(r.Timestamp),
		Buckets:   buckets,
	}
}

func ArrayRoutingTableToPB(in []*RoutingTable) []*pb.Snapshot_RoutingTable {
	out := make([]*pb.Snapshot_RoutingTable, 0, len(in))
	for _, p := range in {
		out = append(out, p.ToPB())
	}
	return out
}

func newRoutingTableFromNode(n *core.IpfsNode) *RoutingTable {
	rt := n.DHT.WAN.RoutingTable()
	buckets := rt.DumpBuckets()
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
		c.sink.PushRoutingTable(routing_table)
		time.Sleep(c.opts.Interval)
	}
}
