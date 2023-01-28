package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	Scope = instrumentation.Scope{
		Name:    "libp2p.io/telemetry/monitor",
		Version: "0.0.0",
	}

	KeyPeerID     = attribute.Key("peer_id")
	KeyExportKind = attribute.Key("export_kind")

	AttrExportKindBandwidth  = KeyExportKind.String("bandwidth")
	AttrExportKindEvents     = KeyExportKind.String("events")
	AttrExportKindMetrics    = KeyExportKind.String("metrics")
	AttrExportKindProperties = KeyExportKind.String("properties")
	AttrExportKindSession    = KeyExportKind.String("session")
)

type Metrics struct {
	m metric.Meter

	// Synchronous
	DiscoveredPeers   syncint64.Counter
	RediscoveredPeers syncint64.Counter

	// Asynchronous
	ActivePeers asyncint64.Gauge
}

type PeerTaskMetrics struct {
	CollectCompleted syncint64.Counter
	CollectFailure   syncint64.Counter
}

type ExporterMetrics struct {
	Exports syncint64.Counter
}

func New(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	DiscoveredPeers, err := m.SyncInt64().Counter(
		"monitor.discovered_peers",
		instrument.WithDescription("Number of peers discovered, including duplicates"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	RediscoveredPeers, err := m.SyncInt64().Counter(
		"monitor.rediscovered_peers",
		instrument.WithDescription("Number of peers rediscovered"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	ActivePeers, err := m.AsyncInt64().Gauge(
		"monitor.active_peers",
		instrument.WithDescription("Number of peers currently being monitored"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		m:                 m,
		DiscoveredPeers:   DiscoveredPeers,
		RediscoveredPeers: RediscoveredPeers,
		ActivePeers:       ActivePeers,
	}, nil
}

func (m *Metrics) RegisterCallback(cb func(context.Context)) error {
	instruments := []instrument.Asynchronous{
		m.ActivePeers,
	}
	return m.m.RegisterCallback(instruments, cb)
}

func NewPeerTaskMetrics(meterProvider metric.MeterProvider) (*PeerTaskMetrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	CollectCompleted, err := m.SyncInt64().Counter(
		"monitor.collect_completed",
		instrument.WithDescription("Number of collect completions"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	CollectFailure, err := m.SyncInt64().Counter(
		"monitor.collect_failure",
		instrument.WithDescription("Number of collect failures"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	return &PeerTaskMetrics{
		CollectCompleted: CollectCompleted,
		CollectFailure:   CollectFailure,
	}, nil
}

func NewExportMetrics(meterProvider metric.MeterProvider) (*ExporterMetrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	Exports, err := m.SyncInt64().Counter(
		"monitor.exports",
		instrument.WithDescription("Number of exports"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	return &ExporterMetrics{
		Exports: Exports,
	}, nil
}
