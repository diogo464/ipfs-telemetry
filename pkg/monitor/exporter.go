package monitor

import (
	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p-core/peer"
)

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
