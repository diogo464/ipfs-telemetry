package collector

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/ipfs/go-ipfs/core"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Collector = (*networkCollector)(nil)

type NetworkOptions struct {
	BandwidthByPeerInterval time.Duration
}

type networkCollector struct {
	opts                   NetworkOptions
	node                   *core.IpfsNode
	last_bandwidth_by_peer time.Time
}

func NewNetworkCollector(n *core.IpfsNode, opts NetworkOptions) Collector {
	return &networkCollector{
		opts:                   opts,
		node:                   n,
		last_bandwidth_by_peer: time.Now(),
	}
}

// Close implements Collector
func (*networkCollector) Close() {
}

// Collect implements Collector
func (c *networkCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	collectBandwidthByPeer := false
	if time.Since(c.last_bandwidth_by_peer) > c.opts.BandwidthByPeerInterval {
		collectBandwidthByPeer = true
		c.last_bandwidth_by_peer = time.Now()
	}
	network := newNetworkFromNode(c.node, collectBandwidthByPeer)
	sink.Push(network)
}

func newNetworkFromNode(n *core.IpfsNode, collectBandwidthByPeer bool) *datapoint.Network {
	reporter := n.Reporter
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	var bandwidthByPeer map[peer.ID]metrics.Stats = nil
	if collectBandwidthByPeer {
		bandwidthByPeer = reporter.GetBandwidthByPeer()
	}
	return &datapoint.Network{
		Timestamp:       datapoint.NewTimestamp(),
		Addresses:       n.PeerHost.Addrs(),
		Overall:         reporter.GetBandwidthTotals(),
		StatsByProtocol: reporter.GetBandwidthByProtocol(),
		StatsByPeer:     bandwidthByPeer,
		NumConns:        uint32(info.ConnCount),
		LowWater:        uint32(info.LowWater),
		HighWater:       uint32(info.HighWater),
	}
}
