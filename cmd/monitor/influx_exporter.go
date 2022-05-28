package main

import (
	"encoding/json"
	"time"

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
func (e *InfluxExporter) ExportSessionInfo(p peer.ID, s telemetry.SessionInfo) {
	point := influxdb2.NewPointWithMeasurement("session_info").
		AddField("boot_time", s.BootTime)
	e.writePointWithTime(p, s.Session, time.Now(), point)
}

// ExportSystemInfo implements monitor.Exporter
func (e *InfluxExporter) ExportSystemInfo(p peer.ID, sess telemetry.Session, info telemetry.SystemInfo) {
	point := influxdb2.NewPointWithMeasurement("system_info").
		AddField("os", info.OS).
		AddField("arch", info.Arch).
		AddField("numcpu", info.NumCPU)
	e.writePointWithTime(p, sess, time.Now(), point)
}

// Export implements monitor.Exporter
func (e *InfluxExporter) ExportDatapoints(p peer.ID, sess telemetry.Session, snaps []datapoint.Datapoint) {
	for _, snap := range snaps {
		switch v := snap.(type) {
		case *datapoint.Ping:
			e.exportPing(p, sess, v)
		case *datapoint.Connections:
			e.exportConnections(p, sess, v)
		case *datapoint.RoutingTable:
			e.exportRoutingTable(p, sess, v)
		case *datapoint.Network:
			e.exportNetwork(p, sess, v)
		case *datapoint.Resources:
			e.exportResources(p, sess, v)
		case *datapoint.TraceRoute:
			e.exportTraceRoute(p, sess, v)
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
		case *datapoint.RelayReservation:
			e.exportRelayReservation(p, sess, v)
		case *datapoint.RelayConnection:
			e.exportRelayConnection(p, sess, v)
		case *datapoint.RelayComplete:
			e.exportRelayComplete(p, sess, v)
		case *datapoint.RelayStats:
			e.exportRelayStats(p, sess, v)
		case *datapoint.HolePunch:
			e.exportHolePunch(p, sess, v)
		default:
		}
	}
}

// ExportBandwidth implements monitor.Exporter
func (e *InfluxExporter) ExportBandwidth(p peer.ID, sess telemetry.Session, bw telemetry.Bandwidth) {
	point := influxdb2.NewPointWithMeasurement("bandwidth").
		AddField("upload", bw.UploadRate).
		AddField("download", bw.DownloadRate)
	e.writePointWithTime(p, sess, time.Now(), point)
}

func (e *InfluxExporter) exportPing(p peer.ID, sess telemetry.Session, snap *datapoint.Ping) {
	data, _ := json.Marshal(snap)
	point := influxdb2.NewPointWithMeasurement("ping").
		AddField("data", data)
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportConnections(p peer.ID, sess telemetry.Session, snap *datapoint.Connections) {
	{
		data, _ := json.Marshal(snap.Connections)
		point := influxdb2.NewPointWithMeasurement("connections").
			AddField("data", data).
			AddField("count", len(snap.Connections))
		e.writePoint(p, sess, snap, point)
	}
	{
		streamCounts := make(map[string]uint32)
		for _, conn := range snap.Connections {
			for _, stream := range conn.Streams {
				if stream.Protocol == "" {
					streamCounts["none"] += 1
				} else {
					streamCounts[stream.Protocol] += 1
				}
			}
		}
		for protocol, count := range streamCounts {
			point := influxdb2.NewPointWithMeasurement("streams").
				AddTag("protocol", protocol).
				AddField("count", count)
			e.writePoint(p, sess, snap, point)
		}
	}
}

func (e *InfluxExporter) exportRoutingTable(p peer.ID, sess telemetry.Session, snap *datapoint.RoutingTable) {
	peers := 0
	for _, b := range snap.Buckets {
		peers += len(b)
	}
	data, _ := json.Marshal(snap.Buckets)
	point := influxdb2.NewPointWithMeasurement("routing_table").
		AddField("data", data).
		AddField("buckets", len(snap.Buckets)).
		AddField("peers", peers)
	e.writePoint(p, sess, snap, point)
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
		AddField("cpu_process", snap.CpuProcess).
		AddField("cpu_system", snap.CpuSystem).
		AddField("memory_used", snap.MemoryUsed).
		AddField("memory_free", snap.MemoryFree).
		AddField("memory_total", snap.MemoryTotal).
		AddField("goroutines", snap.Goroutines)
	e.writePoint(p, sess, snap, point)
}

func (e *InfluxExporter) exportTraceRoute(p peer.ID, sess telemetry.Session, snap *datapoint.TraceRoute) {
	data, _ := json.Marshal(snap)
	point := influxdb2.NewPointWithMeasurement("traceroute").
		AddField("data", data)
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
	exportWithDirection(snap.MessagesOut, "out")
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

func (e *InfluxExporter) exportRelayReservation(p peer.ID, sess telemetry.Session, v *datapoint.RelayReservation) {
	point := influxdb2.NewPointWithMeasurement("relay_reservation").
		AddField("initiator", v.Peer.String())
	e.writePoint(p, sess, v, point)
}

func (e *InfluxExporter) exportRelayConnection(p peer.ID, sess telemetry.Session, v *datapoint.RelayConnection) {
	point := influxdb2.NewPointWithMeasurement("relay_connection").
		AddField("initiator", v.Initiator.String()).
		AddField("target", v.Target.String())
	e.writePoint(p, sess, v, point)
}

func (e *InfluxExporter) exportRelayComplete(p peer.ID, sess telemetry.Session, v *datapoint.RelayComplete) {
	point := influxdb2.NewPointWithMeasurement("relay_complete").
		AddField("duration", v.Duration.Nanoseconds()).
		AddField("initiator", v.Initiator.String()).
		AddField("target", v.Target.String()).
		AddField("bytes_released", v.BytesRelayed)
	e.writePoint(p, sess, v, point)
}

func (e *InfluxExporter) exportRelayStats(p peer.ID, sess telemetry.Session, v *datapoint.RelayStats) {
	point := influxdb2.NewPointWithMeasurement("relay_stats").
		AddField("reservations", v.Reservations).
		AddField("connections", v.Connections).
		AddField("bytes_relayed", v.BytesRelayed).
		AddField("active_connections", v.ActiveConnections)
	e.writePoint(p, sess, v, point)
}

func (e *InfluxExporter) exportHolePunch(p peer.ID, sess telemetry.Session, v *datapoint.HolePunch) {
	point := influxdb2.NewPointWithMeasurement("holepunch").
		AddField("success", v.Success).
		AddField("failure", v.Failure)
	e.writePoint(p, sess, v, point)
}

func (e *InfluxExporter) writePoint(p peer.ID, sess telemetry.Session, snap datapoint.Datapoint, point *write.Point) {
	e.writePointWithTime(p, sess, snap.GetTimestamp(), point)
}

func (e *InfluxExporter) writePointWithTime(p peer.ID, sess telemetry.Session, t time.Time, point *write.Point) {
	point.AddTag("peer", p.Pretty())
	point.AddTag("session", sess.String())
	point.SetTime(t)
	e.writer.WritePoint(point)
}
