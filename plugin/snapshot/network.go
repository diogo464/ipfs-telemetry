package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"github.com/ipfs/go-ipfs/core"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/protocol"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Network struct {
	Timestamp   time.Time `json:"timestamp"`
	Overall     metrics.Stats
	PerProtocol map[protocol.ID]metrics.Stats
	NumConns    uint32 `json:"numconns"`
	LowWater    uint32 `json:"lowwater"`
	HighWater   uint32 `json:"highwater"`
}

func (n *Network) ToPB() *pb.Snapshot_Network {
	byprotocol := make(map[string]*pb.Snapshot_Network_Stats)
	for k, v := range n.PerProtocol {
		byprotocol[string(k)] = metricStatsToPB(&v)
	}

	return &pb.Snapshot_Network{
		Timestamp:       timestamppb.New(n.Timestamp),
		StatsOverall:    metricStatsToPB(&n.Overall),
		StatsByProtocol: byprotocol,
		NumConns:        n.NumConns,
		LowWater:        n.LowWater,
		HighWater:       n.HighWater,
	}
}

func ArrayNetworkToPB(in []*Network) []*pb.Snapshot_Network {
	out := make([]*pb.Snapshot_Network, 0, len(in))
	for _, p := range in {
		out = append(out, p.ToPB())
	}
	return out
}

func NewNetworkFromNode(n *core.IpfsNode) *Network {
	reporter := n.Reporter
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	return &Network{
		Timestamp:   time.Now().UTC(),
		Overall:     reporter.GetBandwidthTotals(),
		PerProtocol: reporter.GetBandwidthByProtocol(),
		NumConns:    uint32(info.ConnCount),
		LowWater:    uint32(info.LowWater),
		HighWater:   uint32(info.HighWater),
	}
}

type NetworkOptions struct {
	Interval time.Duration
}

type NetworkCollector struct {
	opts NetworkOptions
	sink Sink
	node *core.IpfsNode
}

func NewNetworkCollector(n *core.IpfsNode, sink Sink, opts NetworkOptions) *NetworkCollector {
	return &NetworkCollector{opts: opts, sink: sink, node: n}
}

func (c *NetworkCollector) Run() {
	for {
		network := NewNetworkFromNode(c.node)
		c.sink.PushNetwork(network)
		time.Sleep(c.opts.Interval)
	}
}
