package monitor

import (
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"git.d464.sh/adc/telemetry/pkg/telemetry/datapoint"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Exporter = (*NullExporter)(nil)

type Exporter interface {
	ExportSessionInfo(peer.ID, telemetry.SessionInfo)
	ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo)
	ExportDatapoints(peer.ID, telemetry.Session, []datapoint.Datapoint)
	ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth)
}

type fnExporter struct {
	fn func(peer.ID, telemetry.Session, []datapoint.Datapoint)
}

func (e *fnExporter) ExportSessionInfo(peer.ID, telemetry.SessionInfo)                  {}
func (e *fnExporter) ExportSystemInfo(peer.ID, telemetry.Session, telemetry.SystemInfo) {}
func (e *fnExporter) ExportDatapoints(peer peer.ID, session telemetry.Session, datapoints []datapoint.Datapoint) {
	e.fn(peer, session, datapoints)
}
func (e *fnExporter) ExportBandwidth(peer.ID, telemetry.Session, telemetry.Bandwidth) {}

func NewExporterFn(fn func(peer.ID, telemetry.Session, []datapoint.Datapoint)) Exporter {
	return &fnExporter{fn}
}

type NullExporter struct{}

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
