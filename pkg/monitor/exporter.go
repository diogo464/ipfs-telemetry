package monitor

import (
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Exporter = (*NullExporter)(nil)

type Exporter interface {
	ExportSessionInfo(peer.ID, telemetry.SessionInfo)
	ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo)
	ExportSnapshots(peer.ID, telemetry.Session, []snapshot.Snapshot)
	ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth)
}

type fnExporter struct {
	fn func(peer.ID, telemetry.Session, []snapshot.Snapshot)
}

func (e *fnExporter) ExportSessionInfo(peer.ID, telemetry.SessionInfo)                  {}
func (e *fnExporter) ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo) {}
func (e *fnExporter) ExportSnapshots(peer peer.ID, session telemetry.Session, snapshots []snapshot.Snapshot) {
	e.fn(peer, session, snapshots)
}
func (e *fnExporter) ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth) {}

func NewExporterFn(fn func(peer.ID, telemetry.Session, []snapshot.Snapshot)) Exporter {
	return &fnExporter{fn}
}

type NullExporter struct{}

// ExportBandwidth implements Exporter
func (*NullExporter) ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth) {
}

// ExportSessionInfo implements Exporter
func (*NullExporter) ExportSessionInfo(peer.ID, telemetry.SessionInfo) {
}

// ExportSnapshots implements Exporter
func (*NullExporter) ExportSnapshots(peer.ID, telemetry.Session, []snapshot.Snapshot) {
}

// ExportSystemInfo implements Exporter
func (*NullExporter) ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo) {
}
