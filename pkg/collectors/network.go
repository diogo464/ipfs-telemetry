package collectors

import (
	"context"
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/ipfs/kubo/core"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

var _ telemetry.Collector = (*networkCollector)(nil)

type NetworkOptions struct {
	BandwidthByPeerInterval time.Duration
}

type networkCollector struct {
	opts                   NetworkOptions
	node                   *core.IpfsNode
	last_bandwidth_by_peer time.Time
}

func Network(n *core.IpfsNode, opts NetworkOptions) telemetry.Collector {
	return &networkCollector{
		opts:                   opts,
		node:                   n,
		last_bandwidth_by_peer: time.Now(),
	}
}

// Descriptor implements telemetry.Collector
func (*networkCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.NetworkName,
	}
}

// Open implements telemetry.Collector
func (*networkCollector) Open() {
}

// Close implements Collector
func (*networkCollector) Close() {
}

// Collect implements Collector
func (c *networkCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	collectBandwidthByPeer := false
	if time.Since(c.last_bandwidth_by_peer) > c.opts.BandwidthByPeerInterval {
		collectBandwidthByPeer = true
		c.last_bandwidth_by_peer = time.Now()
	}
	network := newNetworkFromNode(c.node, collectBandwidthByPeer)
	return datapoint.NetworkSerialize(network, stream)
}

func newNetworkFromNode(n *core.IpfsNode, collectBandwidthByPeer bool) *datapoint.Network {
	reporter := n.Reporter
	cmgr := n.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
	info := cmgr.GetInfo()
	//var bandwidthByPeer map[peer.ID]metrics.Stats = nil
	//if collectBandwidthByPeer {
	//	bandwidthByPeer = reporter.GetBandwidthByPeer()
	//}
	return &datapoint.Network{
		Timestamp:       datapoint.NewTimestamp(),
		Addresses:       n.PeerHost.Addrs(),
		Overall:         reporter.GetBandwidthTotals(),
		StatsByProtocol: reporter.GetBandwidthByProtocol(),
		//StatsByPeer:     bandwidthByPeer,
		NumConns:  uint32(info.ConnCount),
		LowWater:  uint32(info.LowWater),
		HighWater: uint32(info.HighWater),
	}
}
