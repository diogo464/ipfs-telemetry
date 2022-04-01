package collector

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-ipfs/core"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
)

type NetworkOptions struct {
	Interval time.Duration
}

type networkCollector struct {
	opts NetworkOptions
	sink snapshot.Sink
	node *core.IpfsNode
}

func RunNetworkCollector(n *core.IpfsNode, sink snapshot.Sink, opts NetworkOptions) {
	c := &networkCollector{opts: opts, sink: sink, node: n}
	c.Run()
}

func (c *networkCollector) Run() {
	for {
		network := newNetworkFromNode(c.node)
		c.sink.PushNetwork(network)
		time.Sleep(c.opts.Interval)
	}
}

func newNetworkFromNode(n *core.IpfsNode) *snapshot.Network {
	reporter := n.Reporter
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	return &snapshot.Network{
		Timestamp:   time.Now().UTC(),
		Overall:     reporter.GetBandwidthTotals(),
		PerProtocol: reporter.GetBandwidthByProtocol(),
		NumConns:    uint32(info.ConnCount),
		LowWater:    uint32(info.LowWater),
		HighWater:   uint32(info.HighWater),
	}
}
