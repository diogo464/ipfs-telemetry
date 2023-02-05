package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
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
	DiscoveredPeers   instrument.Int64Counter
	RediscoveredPeers instrument.Int64Counter

	// Asynchronous
	ActivePeers instrument.Int64ObservableGauge
}

type PeerTaskMetrics struct {
	CollectCompleted instrument.Int64Counter
	CollectFailure   instrument.Int64Counter
}

type ExporterMetrics struct {
	Exports instrument.Int64Counter
}

func New(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	DiscoveredPeers, err := m.Int64Counter(
		"monitor.discovered_peers",
		instrument.WithDescription("Number of peers discovered, including duplicates"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	RediscoveredPeers, err := m.Int64Counter(
		"monitor.rediscovered_peers",
		instrument.WithDescription("Number of peers rediscovered"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	ActivePeers, err := m.Int64ObservableGauge(
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

func (m *Metrics) RegisterCallback(cb func(context.Context, metric.Observer) error) error {
	instruments := []instrument.Asynchronous{
		m.ActivePeers,
	}
	_, err := m.m.RegisterCallback(cb, instruments...)
	return err
}

func NewPeerTaskMetrics(meterProvider metric.MeterProvider) (*PeerTaskMetrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	CollectCompleted, err := m.Int64Counter(
		"monitor.collect_completed",
		instrument.WithDescription("Number of collect completions"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	CollectFailure, err := m.Int64Counter(
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

	Exports, err := m.Int64Counter(
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
