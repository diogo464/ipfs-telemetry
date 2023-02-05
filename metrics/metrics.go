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
		Name:    "libp2p.io/telemetry",
		Version: "1.0.0",
	}

	AclScope = instrumentation.Scope{
		Name:    "libp2p.io/telemetry/acl",
		Version: "1.0.0",
	}

	StreamScope = instrumentation.Scope{
		Name:    "libp2p.io/telemetry/stream",
		Version: "1.0.0",
	}

	KeyStreamID   = attribute.Key("stream_id")
	KeyGrpcMethod = attribute.Key("grpc_method")
)

type Metrics struct {
	m metric.Meter

	// Syncronous
	GrpcReqCount     instrument.Int64Counter
	GrpcReqDur       instrument.Int64Histogram
	GrpcStreamSegRet instrument.Int64Histogram

	// Asyncronous
	StreamCount   instrument.Int64ObservableGauge
	PropertyCount instrument.Int64ObservableGauge
	EventCount    instrument.Int64ObservableGauge
}

type AclMetrics struct {
	BlockedRequests instrument.Int64Counter
	AllowedRequests instrument.Int64Counter
}

type StreamMetrics struct {
	m metric.Meter

	// Asyncronous
	UsedSize  instrument.Int64ObservableGauge
	TotalSize instrument.Int64ObservableGauge
}

func NewMetrics(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	GrpcReqCount, err := m.Int64Counter(
		"telemetry.grpc_request_count",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Number of GRPC requests"),
	)
	if err != nil {
		return nil, err
	}

	GrpcReqDur, err := m.Int64Histogram(
		"telemetry.grpc_request_duration",
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription("Histogram of GRPC request duration"),
	)
	if err != nil {
		return nil, err
	}

	GrpcStreamSegRet, err := m.Int64Histogram(
		"telemetry.grpc_stream_segment_return",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Histogram of number of stream segments returned"),
	)
	if err != nil {
		return nil, err
	}

	StreamCount, err := m.Int64ObservableGauge(
		"telemetry.stream_count",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Number of streams"),
	)
	if err != nil {
		return nil, err
	}

	PropertyCount, err := m.Int64ObservableGauge(
		"telemetry.property_count",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Number of properties"),
	)
	if err != nil {
		return nil, err
	}

	EventCount, err := m.Int64ObservableGauge(
		"telemetry.event_count",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Number of events"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		m: m,

		GrpcReqCount:     GrpcReqCount,
		GrpcReqDur:       GrpcReqDur,
		GrpcStreamSegRet: GrpcStreamSegRet,

		StreamCount:   StreamCount,
		PropertyCount: PropertyCount,
		EventCount:    EventCount,
	}, nil
}

func (m *Metrics) RegisterCallback(cb func(context.Context, metric.Observer) error) error {
	instruments := []instrument.Asynchronous{
		m.StreamCount,
		m.PropertyCount,
		m.EventCount,
	}
	_, err := m.m.RegisterCallback(cb, instruments...)
	return err
}

func NewAclMetrics(meterProvider metric.MeterProvider) (*AclMetrics, error) {
	m := meterProvider.Meter(AclScope.Name, metric.WithInstrumentationVersion(AclScope.Version), metric.WithSchemaURL(AclScope.SchemaURL))

	BlockedRequests, err := m.Int64Counter(
		"telemetry.acl.blocked_requests",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Number of blocked requests"),
	)
	if err != nil {
		return nil, err
	}

	AllowedRequests, err := m.Int64Counter(
		"telemetry.acl.allowed_requests",
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription("Number of allowed requests"),
	)
	if err != nil {
		return nil, err
	}

	return &AclMetrics{
		BlockedRequests: BlockedRequests,
		AllowedRequests: AllowedRequests,
	}, nil
}

func NewStreamMetrics(meterProvider metric.MeterProvider) (*StreamMetrics, error) {
	m := meterProvider.Meter(StreamScope.Name, metric.WithInstrumentationVersion(StreamScope.Version), metric.WithSchemaURL(StreamScope.SchemaURL))

	UsedSize, err := m.Int64ObservableGauge(
		"telemetry.stream.used_size",
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("Used size"),
	)
	if err != nil {
		return nil, err
	}

	TotalSize, err := m.Int64ObservableGauge(
		"telemetry.stream.total_size",
		instrument.WithUnit(unit.Bytes),
		instrument.WithDescription("Total size"),
	)
	if err != nil {
		return nil, err
	}

	return &StreamMetrics{
		m: m,

		UsedSize:  UsedSize,
		TotalSize: TotalSize,
	}, nil
}

func (m *StreamMetrics) RegisterCallback(cb func(context.Context, metric.Observer) error) error {
	instruments := []instrument.Asynchronous{
		m.UsedSize,
		m.TotalSize,
	}
	_, err := m.m.RegisterCallback(cb, instruments...)
	return err
}
