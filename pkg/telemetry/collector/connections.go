package collector

import (
	"context"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/libp2p/go-libp2p-core/host"
)

var _ Collector = (*connectionsCollector)(nil)

type connectionsCollector struct {
	h host.Host
}

func NewConnectionsCollector(h host.Host) Collector {
	return &connectionsCollector{
		h: h,
	}
}

// Close implements Collector
func (*connectionsCollector) Close() {
}

// Collect implements Collector
func (c *connectionsCollector) Collect(ctx context.Context, sink datapoint.Sink) {
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
	sink.Push(&datapoint.Connections{
		Timestamp:   datapoint.NewTimestamp(),
		Connections: conns,
	})
}
