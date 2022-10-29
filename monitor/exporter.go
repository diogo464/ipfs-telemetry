package monitor

import (
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p/core/peer"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

var _ (Exporter) = (*noOpExporter)(nil)

type Capture struct {
	Timestamp time.Time
	Data      []byte
}

type Event struct {
	Timestamp time.Time
	Data      []byte
}

type Exporter interface {
	Session(peer.ID, telemetry.Session, []telemetry.CProperty)
	Metrics(peer.ID, telemetry.Session, []*mpb.ResourceMetrics)
	Captures(peer.ID, telemetry.Session, telemetry.CaptureDescriptor, []Capture)
	Events(peer.ID, telemetry.Session, telemetry.EventDescriptor, []Event)
	Bandwidth(peer.ID, telemetry.Bandwidth)
}

type noOpExporter struct{}

func NewNoOpExporter() Exporter {
	return &noOpExporter{}
}

// Session implements Exporter
func (*noOpExporter) Session(peer.ID, telemetry.Session, []telemetry.CProperty) {
}

// Captures implements Exporter
func (*noOpExporter) Captures(peer.ID, telemetry.Session, telemetry.CaptureDescriptor, []Capture) {
}

// Events implements Exporter
func (*noOpExporter) Events(peer.ID, telemetry.Session, telemetry.EventDescriptor, []Event) {
}

// Metrics implements Exporter
func (*noOpExporter) Metrics(peer.ID, telemetry.Session, []*mpb.ResourceMetrics) {
}

// Bandwidth implements Exporter
func (*noOpExporter) Bandwidth(peer.ID, telemetry.Bandwidth) {
}
