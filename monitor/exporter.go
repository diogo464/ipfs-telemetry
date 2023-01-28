package monitor

import (
	"context"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/monitor/metrics"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ (Exporter) = (*noOpExporter)(nil)
var _ (Exporter) = (*observableExporter)(nil)

type Exporter interface {
	Session(peer.ID, telemetry.Session)
	Metrics(peer.ID, telemetry.Session, telemetry.Metrics)
	Properties(peer.ID, telemetry.Session, []telemetry.Property)
	Events(peer.ID, telemetry.Session, telemetry.EventDescriptor, []telemetry.Event)
	Bandwidth(peer.ID, telemetry.Bandwidth)
}

type noOpExporter struct{}

func NewNoOpExporter() Exporter {
	return &noOpExporter{}
}

// Properties implements Exporter
func (*noOpExporter) Properties(peer.ID, telemetry.Session, []telemetry.Property) {
}

// Session implements Exporter
func (*noOpExporter) Session(peer.ID, telemetry.Session) {
}

// Events implements Exporter
func (*noOpExporter) Events(peer.ID, telemetry.Session, telemetry.EventDescriptor, []telemetry.Event) {
}

// Metrics implements Exporter
func (*noOpExporter) Metrics(peer.ID, telemetry.Session, telemetry.Metrics) {
}

// Bandwidth implements Exporter
func (*noOpExporter) Bandwidth(peer.ID, telemetry.Bandwidth) {
}

type observableExporter struct {
	m *metrics.ExporterMetrics
	e Exporter
}

// Bandwidth implements Exporter
func (e *observableExporter) Bandwidth(p peer.ID, b telemetry.Bandwidth) {
	e.m.Exports.Add(context.Background(), 1, metrics.AttrExportKindBandwidth)
	e.e.Bandwidth(p, b)
}

// Events implements Exporter
func (e *observableExporter) Events(p peer.ID, s telemetry.Session, d telemetry.EventDescriptor, ev []telemetry.Event) {
	e.m.Exports.Add(context.Background(), 1, metrics.AttrExportKindEvents)
	e.e.Events(p, s, d, ev)
}

// Metrics implements Exporter
func (e *observableExporter) Metrics(p peer.ID, s telemetry.Session, m telemetry.Metrics) {
	e.m.Exports.Add(context.Background(), 1, metrics.AttrExportKindMetrics)
	e.e.Metrics(p, s, m)
}

// Properties implements Exporter
func (e *observableExporter) Properties(p peer.ID, s telemetry.Session, pp []telemetry.Property) {
	e.m.Exports.Add(context.Background(), 1, metrics.AttrExportKindProperties)
	e.e.Properties(p, s, pp)
}

// Session implements Exporter
func (e *observableExporter) Session(p peer.ID, s telemetry.Session) {
	e.m.Exports.Add(context.Background(), 1, metrics.AttrExportKindSession)
	e.e.Session(p, s)
}
