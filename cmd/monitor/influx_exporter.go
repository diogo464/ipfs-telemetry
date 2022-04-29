package main

import (
	"strconv"

	"github.com/diogo464/telemetry/pkg/monitor"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/libp2p/go-libp2p-core/peer"
)

// *pb.Datapoint_Ping:
// *pb.Datapoint_RoutingTable:
// *pb.Datapoint_Network:
// *pb.Datapoint_Resources:
// *pb.Datapoint_Traceroute:
// *pb.Datapoint_Kademlia:
// *pb.Datapoint_KademliaQuery:
// *pb.Datapoint_Bitswap:
// *pb.Datapoint_Ipns:
// *pb.Datapoint_Storage:

// ping,node=...,session=...,origin=<public ip>,destination=<public ip> duration=...
// routing_table,node=...,session=...,bucket=...,position=... peer=

// network_connectivity,node=...,session=...,

var _ monitor.Exporter = (*InfluxExporter)(nil)

type InfluxExporter struct {
	writer api.WriteAPI
}

func NewInfluxExporter(writer api.WriteAPI) *InfluxExporter {
	return &InfluxExporter{
		writer: writer,
	}
}

func (e *InfluxExporter) Close() {
	e.writer.Flush()
}

// ExportSessionInfo implements monitor.Exporter
func (*InfluxExporter) ExportSessionInfo(peer.ID, telemetry.SessionInfo) {
	panic("unimplemented")
}

// ExportSystemInfo implements monitor.Exporter
func (*InfluxExporter) ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo) {
	panic("unimplemented")
}

// Export implements monitor.Exporter
func (e *InfluxExporter) ExportDatapoints(p peer.ID, sess telemetry.Session, snaps []datapoint.Datapoint) {
	for _, snap := range snaps {
		switch v := snap.(type) {
		case *datapoint.Ping:
			e.exportPing(p, sess, v)
		case *datapoint.RoutingTable:
			e.exportRoutingTable(p, sess, v)
		case *datapoint.Network:
			e.exportNetwork(p, sess, v)
		case *datapoint.Resources:
			e.exportResources(p, sess, v)
		case *datapoint.TraceRoute:
		case *datapoint.Kademlia:
			e.exportKademlia(p, sess, v)
		case *datapoint.KademliaQuery:
			e.exportKademliaQuery(p, sess, v)
		case *datapoint.KademliaHandler:
			e.exportKademliaHandler(p, sess, v)
		case *datapoint.Bitswap:
			e.exportBitswap(p, sess, v)
		case *datapoint.Storage:
			e.exportStorage(p, sess, v)
		case *datapoint.Window:
			e.exportWindow(p, sess, v)
		}
	}
}

// ExportBandwidth implements monitor.Exporter
func (e *InfluxExporter) ExportBandwidth(p peer.ID, sess telemetry.Session, bw telemetry.Bandwidth) {
}

func (e *InfluxExporter) exportPing(p peer.ID, sess telemetry.Session, snap *datapoint.Ping) {
}

func (e *InfluxExporter) exportRoutingTable(p peer.ID, sess telemetry.Session, snap *datapoint.RoutingTable) {
	for bucket_index, bucket := range snap.Buckets {
		point := influxdb2.NewPointWithMeasurement("routing_table").
			AddTag("bucket", strconv.Itoa(bucket_index)).
			AddField("size", len(bucket))
		e.writePoint(p, sess, snap, point)
	}
}

func (e *InfluxExporter) exportNetwork(p peer.ID, sess telemetry.Session, snap *datapoint.Network) {
	{
		point := influxdb2.NewPointWithMeasurement("network_conns").
			AddField("conns", snap.NumConns).
			AddField("low_water", snap.LowWater).
			AddField("high_water", snap.HighWater)
		e.writePoint(p, sess, snap, point)
	}
	{
		for protocol, stats := range snap.StatsByProtocol {
			point := influxdb2.NewPointWithMeasurement("network_stats").
				AddTag("protocol", string(protocol)).
				AddField("total_in", stats.TotalIn).
				AddField("total_out", stats.TotalOut).
				AddField("rate_in", stats.RateIn).
				AddField("rate_out", stats.RateOut)
			e.writePoint(p, sess, snap, point)
		}
	}
}

func (e *InfluxExporter) exportResources(p peer.ID, sess telemetry.Session, snap *datapoint.Resources) {
	point := influxdb2.NewPointWithMeasurement("resources").
		AddField("cpu", snap.CpuUsage).
		AddField("memory_used", snap.MemoryUsed).
		AddField("memory_free", snap.MemoryFree).
		AddField("memory_total", snap.MemoryTotal).
		AddField("goroutines", snap.Goroutines)
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportKademlia(p peer.ID, sess telemetry.Session, snap *datapoint.Kademlia) {
	exportWithDirection := func(in map[datapoint.KademliaMessageType]uint64, dir string) {
		for ty, count := range in {
			point := influxdb2.NewPointWithMeasurement("kademlia").
				AddTag("direction", dir).
				AddTag("type", datapoint.KademliaMessageTypeString[ty]).
				AddField("count", count)
			e.writePoint(p, sess, snap, point)
		}
	}
	exportWithDirection(snap.MessagesIn, "in")
	exportWithDirection(snap.MessagesIn, "out")
}

func (e *InfluxExporter) exportKademliaQuery(p peer.ID, sess telemetry.Session, snap *datapoint.KademliaQuery) {
	point := influxdb2.NewPointWithMeasurement("kademlia_query").
		AddTag("remote_peer", snap.Peer.Pretty()).
		AddField("type", datapoint.KademliaMessageTypeString[snap.QueryType]).
		AddField("duration", snap.Duration.Nanoseconds())
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportKademliaHandler(p peer.ID, sess telemetry.Session, snap *datapoint.KademliaHandler) {
	point := influxdb2.NewPointWithMeasurement("kademlia_handler").
		AddTag("type", datapoint.KademliaMessageTypeString[snap.HandlerType]).
		AddField("handler", snap.HandlerDuration.Nanoseconds()).
		AddField("write", snap.WriteDuration.Nanoseconds()).
		AddField("total", (snap.WriteDuration + snap.HandlerDuration).Nanoseconds())
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportBitswap(p peer.ID, sess telemetry.Session, snap *datapoint.Bitswap) {
	point := influxdb2.NewPointWithMeasurement("bitswap").
		AddField("messages_in", snap.MessagesIn).
		AddField("messages_out", snap.MessagesOut).
		AddField("discovery_succeeded", snap.DiscoverySucceeded).
		AddField("discovery_failed", snap.DiscoveryFailed)
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportStorage(p peer.ID, sess telemetry.Session, snap *datapoint.Storage) {
	point := influxdb2.NewPointWithMeasurement("storage").
		AddField("storage_used", snap.StorageUsed).
		AddField("storage_total", snap.StorageTotal).
		AddField("num_objects", snap.NumObjects)
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportWindow(p peer.ID, sess telemetry.Session, snap *datapoint.Window) {
	{
		point := influxdb2.NewPointWithMeasurement("window_count")
		for k, v := range snap.DatapointCount {
			point.AddField(k, v)
		}
		e.writePoint(p, sess, snap, point)
	}
	{
		point := influxdb2.NewPointWithMeasurement("window_memory")
		for k, v := range snap.DatapointMemory {
			point.AddField(k, v)
		}
		e.writePoint(p, sess, snap, point)
	}
}

func (e *InfluxExporter) writePoint(p peer.ID, sess telemetry.Session, snap datapoint.Datapoint, point *write.Point) {
	point.AddTag("peer", p.Pretty())
	point.AddTag("session", sess.String())
	point.SetTime(snap.GetTimestamp())
	e.writer.WritePoint(point)
}
