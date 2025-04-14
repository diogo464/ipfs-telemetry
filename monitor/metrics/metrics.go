package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
	DiscoveredPeers   metric.Int64Counter
	RediscoveredPeers metric.Int64Counter

	// Asynchronous
	ActivePeers metric.Int64ObservableGauge
}

type PeerTaskMetrics struct {
	CollectCompleted metric.Int64Counter
	CollectFailure   metric.Int64Counter
}

type ExporterMetrics struct {
	Exports metric.Int64Counter
}

func New(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	DiscoveredPeers, err := m.Int64Counter(
		"monitor.discovered_peers",
		metric.WithDescription("Number of peers discovered, including duplicates"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	RediscoveredPeers, err := m.Int64Counter(
		"monitor.rediscovered_peers",
		metric.WithDescription("Number of peers rediscovered"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	ActivePeers, err := m.Int64ObservableGauge(
		"monitor.active_peers",
		metric.WithDescription("Number of peers currently being monitored"),
		metric.WithUnit("1"),
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
	instruments := []metric.Asynchronous{
		m.ActivePeers,
	}
	_, err := m.m.RegisterCallback(cb, instruments...)
	return err
}

func NewPeerTaskMetrics(meterProvider metric.MeterProvider) (*PeerTaskMetrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	CollectCompleted, err := m.Int64Counter(
		"monitor.collect_completed",
		metric.WithDescription("Number of collect completions"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	CollectFailure, err := m.Int64Counter(
		"monitor.collect_failure",
		metric.WithDescription("Number of collect failures"),
		metric.WithUnit("1"),
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
		metric.WithDescription("Number of exports"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	return &ExporterMetrics{
		Exports: Exports,
	}, nil
}
