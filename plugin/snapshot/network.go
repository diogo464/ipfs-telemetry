package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"github.com/ipfs/go-ipfs/core"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Network struct {
	Timestamp time.Time `json:"timestamp"`
	TotalIn   uint64    `json:"totalin"`
	TotalOut  uint64    `json:"totalout"`
	RateIn    uint64    `json:"ratein"`
	RateOut   uint64    `json:"rateout"`
	NumConns  uint32    `json:"numconns"`
	LowWater  uint32    `json:"lowwater"`
	HighWater uint32    `json:"highwater"`
}

func (n *Network) ToPB() *pb.Snapshot_Network {
	return &pb.Snapshot_Network{
		Timestamp: timestamppb.New(n.Timestamp),
		TotalIn:   n.TotalIn,
		TotalOut:  n.TotalOut,
		RateIn:    n.RateIn,
		RateOut:   n.RateOut,
		NumConns:  n.NumConns,
		LowWater:  n.LowWater,
		HighWater: n.HighWater,
	}
}

func ArrayNetworkToPB(in []*Network) []*pb.Snapshot_Network {
	out := make([]*pb.Snapshot_Network, 0, len(in))
	for _, p := range in {
		out = append(out, p.ToPB())
	}
	return out
}

func NewNetwork(ti uint64, to uint64, ri uint64, ro uint64, nc uint32, lw uint32, hw uint32) *Network {
	return &Network{
		TotalIn:   ti,
		TotalOut:  to,
		RateIn:    ri,
		RateOut:   ro,
		NumConns:  nc,
		LowWater:  lw,
		HighWater: hw,
	}
}

func NewNetworkFromNode(n *core.IpfsNode) *Network {
	reporter := n.Reporter
	totals := reporter.GetBandwidthTotals()
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	return NewNetwork(uint64(totals.TotalIn), uint64(totals.TotalOut), uint64(totals.RateIn), uint64(totals.RateOut), uint32(info.ConnCount), uint32(info.LowWater), uint32(info.HighWater))
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
