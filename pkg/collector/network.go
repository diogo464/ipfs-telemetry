package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-ipfs/core"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
)

type NetworkOptions struct {
	Interval time.Duration
}

type networkCollector struct {
	ctx  context.Context
	opts NetworkOptions
	sink snapshot.Sink
	node *core.IpfsNode
}

func RunNetworkCollector(ctx context.Context, n *core.IpfsNode, sink snapshot.Sink, opts NetworkOptions) {
	c := &networkCollector{ctx: ctx, opts: opts, sink: sink, node: n}
	c.Run()
}

func (c *networkCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)

LOOP:
	for {
		select {
		case <-ticker.C:
			network := newNetworkFromNode(c.node)
			c.sink.Push(network)
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

func newNetworkFromNode(n *core.IpfsNode) *snapshot.Network {
	reporter := n.Reporter
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	return &snapshot.Network{
		Timestamp:       snapshot.NewTimestamp(),
		Overall:         reporter.GetBandwidthTotals(),
		StatsByProtocol: reporter.GetBandwidthByProtocol(),
		StatsByPeer:     reporter.GetBandwidthByPeer(),
		NumConns:        uint32(info.ConnCount),
		LowWater:        uint32(info.LowWater),
		HighWater:       uint32(info.HighWater),
	}
}
