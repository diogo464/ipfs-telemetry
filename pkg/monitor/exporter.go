package monitor

import (
	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Exporter = (*NullExporter)(nil)
var _ Exporter = (*fnExporter)(nil)

type Exporter interface {
	ExportSessionInfo(peer.ID, telemetry.SessionInfo)
	ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo)
	ExportDatapoints(peer.ID, telemetry.Session, []datapoint.Datapoint)
	ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth)
	ExportProviderRecords(peer.ID, telemetry.Session, []telemetry.ProviderRecord)
}

type fnExporter struct {
	fn func(peer.ID, telemetry.Session, []datapoint.Datapoint)
}

func (e *fnExporter) ExportSessionInfo(peer.ID, telemetry.SessionInfo)                  {}
func (e *fnExporter) ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo) {}
func (e *fnExporter) ExportDatapoints(peer peer.ID, session telemetry.Session, datapoints []datapoint.Datapoint) {
	e.fn(peer, session, datapoints)
}
func (e *fnExporter) ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth)              {}
func (e *fnExporter) ExportProviderRecords(peer.ID, telemetry.Session, []telemetry.ProviderRecord) {}

func NewExporterFn(fn func(peer.ID, telemetry.Session, []datapoint.Datapoint)) Exporter {
	return &fnExporter{fn}
}

type NullExporter struct{}

// ExportProviderRecords implements Exporter
func (*NullExporter) ExportProviderRecords(peer.ID, telemetry.Session, []telemetry.ProviderRecord) {
}

// ExportBandwidth implements Exporter
func (*NullExporter) ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth) {
}

// ExportSessionInfo implements Exporter
func (*NullExporter) ExportSessionInfo(peer.ID, telemetry.SessionInfo) {
}

// ExportDatapoints implements Exporter
func (*NullExporter) ExportDatapoints(peer.ID, telemetry.Session, []datapoint.Datapoint) {
}

// ExportSystemInfo implements Exporter
func (*NullExporter) ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo) {
}
