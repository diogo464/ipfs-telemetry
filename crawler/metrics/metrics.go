package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	Scope = instrumentation.Scope{
		Name:    "libp2p.io/telemetry/crawler",
		Version: "0.0.0",
	}
)

type Metrics struct {
	m metric.Meter

	PeersCurrentCrawl          instrument.Int64ObservableGauge
	PeersLastCrawl             instrument.Int64ObservableGauge
	PeersTelemetryCurrentCrawl instrument.Int64ObservableGauge
	PeersTelemetryLastCrawl    instrument.Int64ObservableGauge
	ErrorsCurrentCrawl         instrument.Int64ObservableGauge
	ErrorsLastCrawl            instrument.Int64ObservableGauge
	CompletedCrawls            instrument.Int64ObservableGauge
}

func New(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	PeersCurrentCrawl, err := m.Int64ObservableGauge(
		"crawler.peers",
		instrument.WithDescription("Number unique peers found in this crawl"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	PeersLastCrawl, err := m.Int64ObservableGauge(
		"crawler.peers_last_crawl",
		instrument.WithDescription("Number of unique peers found in the last crawl"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	PeersTelemetryCurrentCrawl, err := m.Int64ObservableGauge(
		"crawler.peers_telemetry",
		instrument.WithDescription("Number of unique peers found in this crawl supporting the telemetry protocol"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	PeersTelemetryLastCrawl, err := m.Int64ObservableGauge(
		"crawler.peers_telemetry_last_crawl",
		instrument.WithDescription("Number of unique peers found in the last crawl supporting the telemetry protocol"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	ErrorsCurrentCrawl, err := m.Int64ObservableGauge(
		"crawler.errors",
		instrument.WithDescription("Number of errors in this crawl"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	ErrorsLastCrawl, err := m.Int64ObservableGauge(
		"crawler.errors_last_crawl",
		instrument.WithDescription("Number of errros in the last crawl"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	CompletedCrawls, err := m.Int64ObservableGauge(
		"crawler.completed_crawls",
		instrument.WithDescription("Number of completed crawls"),
		instrument.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		m: m,

		PeersCurrentCrawl:          PeersCurrentCrawl,
		PeersLastCrawl:             PeersLastCrawl,
		PeersTelemetryCurrentCrawl: PeersTelemetryCurrentCrawl,
		PeersTelemetryLastCrawl:    PeersTelemetryLastCrawl,
		ErrorsCurrentCrawl:         ErrorsCurrentCrawl,
		ErrorsLastCrawl:            ErrorsLastCrawl,
		CompletedCrawls:            CompletedCrawls,
	}, nil
}

func (m *Metrics) RegisterCallback(cb func(context.Context, metric.Observer) error) error {
	instruments := []instrument.Asynchronous{
		m.PeersCurrentCrawl,
		m.PeersLastCrawl,
		m.PeersTelemetryCurrentCrawl,
		m.PeersTelemetryLastCrawl,
		m.ErrorsCurrentCrawl,
		m.ErrorsLastCrawl,
		m.CompletedCrawls,
	}
	_, err := m.m.RegisterCallback(cb, instruments...)
	return err
}
