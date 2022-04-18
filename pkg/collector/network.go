package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/ipfs/go-ipfs/core"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
)

type NetworkOptions struct {
	Interval                time.Duration
	BandwidthByPeerInterval time.Duration
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
	last_bandwidth_by_peer := time.Now()

LOOP:
	for {
		select {
		case <-ticker.C:
			collectBandwidthByPeer := false
			if time.Since(last_bandwidth_by_peer) > c.opts.BandwidthByPeerInterval {
				collectBandwidthByPeer = true
				last_bandwidth_by_peer = time.Now()
			}
			network := newNetworkFromNode(c.node, collectBandwidthByPeer)
			c.sink.Push(network)
		case <-c.ctx.Done():
			break LOOP
		}
	}
}

func newNetworkFromNode(n *core.IpfsNode, collectBandwidthByPeer bool) *snapshot.Network {
	reporter := n.Reporter
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	var bandwidthByPeer map[peer.ID]metrics.Stats = nil
	if collectBandwidthByPeer {
		bandwidthByPeer = reporter.GetBandwidthByPeer()
	}
	return &snapshot.Network{
		Timestamp:       snapshot.NewTimestamp(),
		Addresses:       n.PeerHost.Addrs(),
		Overall:         reporter.GetBandwidthTotals(),
		StatsByProtocol: reporter.GetBandwidthByProtocol(),
		StatsByPeer:     bandwidthByPeer,
		NumConns:        uint32(info.ConnCount),
		LowWater:        uint32(info.LowWater),
		HighWater:       uint32(info.HighWater),
	}
}
