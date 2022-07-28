package collectors

import (
	"context"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/host"
)

var _ telemetry.Collector = (*connectionsCollector)(nil)

type connectionsCollector struct {
	h host.Host
}

func Connections(h host.Host) telemetry.Collector {
	return &connectionsCollector{
		h: h,
	}
}

// Descriptor implements telemetry.Collector
func (*connectionsCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.ConnectionsName,
	}
}

// Open implements telemetry.Collector
func (*connectionsCollector) Open() {
}

// Close implements Collector
func (*connectionsCollector) Close() {
}

// Collect implements Collector
func (c *connectionsCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	networkConns := c.h.Network().Conns()
	conns := make([]datapoint.Connection, 0, len(networkConns))
	for _, conn := range networkConns {
		streams := make([]datapoint.Stream, 0, len(conn.GetStreams()))
		for _, stream := range conn.GetStreams() {
			streams = append(streams, datapoint.Stream{
				Protocol:  string(stream.Protocol()),
				Opened:    stream.Stat().Opened,
				Direction: stream.Stat().Direction,
			})
		}
		conns = append(conns, datapoint.Connection{
			ID:      conn.RemotePeer(),
			Addr:    conn.RemoteMultiaddr(),
			Latency: c.h.Network().Peerstore().LatencyEWMA(conn.RemotePeer()),
			Opened:  conn.Stat().Opened,
			Streams: streams,
		})
	}
	dp := &datapoint.Connections{
		Timestamp:   datapoint.NewTimestamp(),
		Connections: conns,
	}
	return datapoint.ConnectionsSerialize(dp, stream)
}
