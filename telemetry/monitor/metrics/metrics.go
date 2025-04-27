package metrics

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
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
	KeyReason     = attribute.Key("reason")
	KeyOperation  = attribute.Key("operation")

	AttrPeerTaskOp_CreateClient  = KeyOperation.String("create_client")
	AttrPeerTaskOp_GetSession    = KeyOperation.String("get_session")
	AttrPeerTaskOp_ExportSession = KeyOperation.String("")

	AttrExportKindBandwidth  = KeyExportKind.String("bandwidth")
	AttrExportKindEvents     = KeyExportKind.String("events")
	AttrExportKindMetrics    = KeyExportKind.String("metrics")
	AttrExportKindProperties = KeyExportKind.String("properties")
	AttrExportKindSession    = KeyExportKind.String("session")

	histogramBucketsMs = []float64{0.01, 0.05, 0.1, 0.3, 0.6, 0.8, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}

	unitCount = "{count}"
	unitMs    = "ms"
)

type Metrics struct {
	m metric.Meter

	discoveredPeers   metric.Int64Counter
	rediscoveredPeers metric.Int64Counter
	activePeers       metric.Int64Gauge
}

type PeerTaskMetrics struct {
	peerId               peer.ID
	collectSuccess       metric.Int64Counter
	collectFailure       metric.Int64Counter
	collectDuration      metric.Float64Histogram
	createClientDuration metric.Float64Histogram
}

type ExporterMetrics struct {
	Exports metric.Int64Counter
}

func New(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	discoveredPeers, err := m.Int64Counter(
		"monitor.discovered",
		metric.WithDescription("Total number of newly discovered peers. Peers that previously failed and have been rediscovered will count as a new discovery."),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	rediscoveredPeers, err := m.Int64Counter(
		"monitor.rediscovered",
		metric.WithDescription("Total number of peers rediscovered. The monitor was already tracking this peers."),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	activePeers, err := m.Int64Gauge(
		"monitor.active",
		metric.WithDescription("Total number of currently active/tracked peers by the monitor."),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		m:                 m,
		discoveredPeers:   discoveredPeers,
		rediscoveredPeers: rediscoveredPeers,
		activePeers:       activePeers,
	}, nil
}

func (m *Metrics) RecordDiscover(peerId peer.ID) {
	m.discoveredPeers.Add(context.Background(), 1, metric.WithAttributes(KeyPeerID.String(peerId.String())))
}

func (m *Metrics) RecordRediscover(peerId peer.ID) {
	m.rediscoveredPeers.Add(context.Background(), 1, metric.WithAttributes(KeyPeerID.String(peerId.String())))
}

func (m *Metrics) RecordActivePeers(active int) {
	m.activePeers.Record(context.Background(), int64(active))
}

func NewPeerTaskMetrics(meterProvider metric.MeterProvider, peerId peer.ID) (*PeerTaskMetrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	collectSuccess, err := m.Int64Counter(
		"monitor.peer.collect_success",
		metric.WithDescription("Total number of successful telemetry collection attempts"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	collectFailure, err := m.Int64Counter(
		"monitor.peer.collect_failure",
		metric.WithDescription("Total number of failed telemetry collection attempts"),
		metric.WithUnit(unitCount),
	)
	if err != nil {
		return nil, err
	}

	collectDuration, err := m.Float64Histogram(
		"monitor.peer.collect_duration",
		metric.WithDescription("Duration of telemetry collection operations"),
		metric.WithUnit(unitMs),
		metric.WithExplicitBucketBoundaries(histogramBucketsMs...),
	)
	if err != nil {
		return nil, err
	}

	createClientLatency, err := m.Float64Histogram(
		"monitor.peer.create_client_duration",
		metric.WithDescription("Duration of client creation operations"),
		metric.WithUnit(unitMs),
		metric.WithExplicitBucketBoundaries(histogramBucketsMs...),
	)

	return &PeerTaskMetrics{
		peerId:               peerId,
		collectSuccess:       collectSuccess,
		collectFailure:       collectFailure,
		collectDuration:      collectDuration,
		createClientDuration: createClientLatency,
	}, nil
}

func (m *PeerTaskMetrics) RecordCollectSuccess(ctx context.Context, dur time.Duration) {
	options := metric.WithAttributes(KeyPeerID.String(m.peerId.String()))
	m.collectSuccess.Add(ctx, 1, options)
	m.collectDuration.Record(ctx, durationToMillis(dur), options)
}

func (m *PeerTaskMetrics) RecordCollectFailure(ctx context.Context, reason string) {
	m.collectFailure.Add(ctx, 1, metric.WithAttributes(
		KeyPeerID.String(m.peerId.String()),
		KeyReason.String(reason),
	))
}

func (m *PeerTaskMetrics) RecordCreateClientDuration(ctx context.Context, dur time.Duration) {
	m.createClientDuration.Record(ctx, durationToMillis(dur), metric.WithAttributes(
		KeyPeerID.String(m.peerId.String()),
	))
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

func durationToMillis(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}
