package metrics

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	Scope = instrumentation.Scope{
		Name:    "libp2p.io/telemetry/crawler",
		Version: "0.0.0",
	}

	unitCount = "{count}"
)

type Metrics struct {
	m metric.Meter

	PeersCurrentCrawl          metric.Int64ObservableGauge
	PeersLastCrawl             metric.Int64ObservableGauge
	PeersTelemetryCurrentCrawl metric.Int64ObservableGauge
	PeersTelemetryLastCrawl    metric.Int64ObservableGauge
	ErrorsCurrentCrawl         metric.Int64ObservableGauge
	ErrorsLastCrawl            metric.Int64ObservableGauge
	CompletedCrawls            metric.Int64ObservableGauge
}

func New(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	PeersCurrentCrawl, err := m.Int64ObservableGauge(
		"crawler.peers",
		metric.WithDescription("Number unique peers found in this crawl"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	PeersLastCrawl, err := m.Int64ObservableGauge(
		"crawler.peers_last_crawl",
		metric.WithDescription("Number of unique peers found in the last crawl"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	PeersTelemetryCurrentCrawl, err := m.Int64ObservableGauge(
		"crawler.peers_telemetry",
		metric.WithDescription("Number of unique peers found in this crawl supporting the telemetry protocol"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	PeersTelemetryLastCrawl, err := m.Int64ObservableGauge(
		"crawler.peers_telemetry_last_crawl",
		metric.WithDescription("Number of unique peers found in the last crawl supporting the telemetry protocol"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	ErrorsCurrentCrawl, err := m.Int64ObservableGauge(
		"crawler.errors",
		metric.WithDescription("Number of errors in this crawl"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	ErrorsLastCrawl, err := m.Int64ObservableGauge(
		"crawler.errors_last_crawl",
		metric.WithDescription("Number of errros in the last crawl"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	CompletedCrawls, err := m.Int64ObservableGauge(
		"crawler.completed_crawls",
		metric.WithDescription("Number of completed crawls"),
		metric.WithUnit(unitCount),
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
	_, err := m.m.RegisterCallback(cb,
		m.PeersCurrentCrawl,
		m.PeersLastCrawl,
		m.PeersTelemetryCurrentCrawl,
		m.PeersTelemetryLastCrawl,
		m.ErrorsCurrentCrawl,
		m.ErrorsLastCrawl,
		m.CompletedCrawls,
	)
	return err
}
