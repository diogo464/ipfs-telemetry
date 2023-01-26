package monitor

import (
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p/core/peer"
)

var _ (Exporter) = (*noOpExporter)(nil)

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
