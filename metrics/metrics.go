package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
	GrpcReqCount     metric.Int64Counter
	GrpcReqDur       metric.Int64Histogram
	GrpcStreamSegRet metric.Int64Histogram

	// Asyncronous
	StreamCount   metric.Int64ObservableGauge
	PropertyCount metric.Int64ObservableGauge
	EventCount    metric.Int64ObservableGauge
}

type AclMetrics struct {
	BlockedRequests metric.Int64Counter
	AllowedRequests metric.Int64Counter
}

type StreamMetrics struct {
	m metric.Meter

	// Asyncronous
	UsedSize  metric.Int64ObservableGauge
	TotalSize metric.Int64ObservableGauge
}

func NewMetrics(meterProvider metric.MeterProvider) (*Metrics, error) {
	m := meterProvider.Meter(Scope.Name, metric.WithInstrumentationVersion(Scope.Version), metric.WithSchemaURL(Scope.SchemaURL))

	GrpcReqCount, err := m.Int64Counter(
		"telemetry.grpc_request_count",
		metric.WithUnit("1"),
		metric.WithDescription("Number of GRPC requests"),
	)
	if err != nil {
		return nil, err
	}

	GrpcReqDur, err := m.Int64Histogram(
		"telemetry.grpc_request_duration",
		metric.WithUnit("ms"),
		metric.WithDescription("Histogram of GRPC request duration"),
	)
	if err != nil {
		return nil, err
	}

	GrpcStreamSegRet, err := m.Int64Histogram(
		"telemetry.grpc_stream_segment_return",
		metric.WithUnit("1"),
		metric.WithDescription("Histogram of number of stream segments returned"),
	)
	if err != nil {
		return nil, err
	}

	StreamCount, err := m.Int64ObservableGauge(
		"telemetry.stream_count",
		metric.WithUnit("1"),
		metric.WithDescription("Number of streams"),
	)
	if err != nil {
		return nil, err
	}

	PropertyCount, err := m.Int64ObservableGauge(
		"telemetry.property_count",
		metric.WithUnit("1"),
		metric.WithDescription("Number of properties"),
	)
	if err != nil {
		return nil, err
	}

	EventCount, err := m.Int64ObservableGauge(
		"telemetry.event_count",
		metric.WithUnit("1"),
		metric.WithDescription("Number of events"),
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
	_, err := m.m.RegisterCallback(cb, m.StreamCount, m.PropertyCount, m.EventCount)
	return err
}

func NewAclMetrics(meterProvider metric.MeterProvider) (*AclMetrics, error) {
	fmt.Println("KEKAA")
	m := meterProvider.Meter(AclScope.Name, metric.WithInstrumentationVersion(AclScope.Version), metric.WithSchemaURL(AclScope.SchemaURL))
	fmt.Println("KEKB")

	BlockedRequests, err := m.Int64Counter(
		"telemetry.acl.blocked_requests",
		metric.WithUnit("1"),
		metric.WithDescription("Number of blocked requests"),
	)
	if err != nil {
		return nil, err
	}

	AllowedRequests, err := m.Int64Counter(
		"telemetry.acl.allowed_requests",
		metric.WithUnit("1"),
		metric.WithDescription("Number of allowed requests"),
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
		metric.WithUnit("By"),
		metric.WithDescription("Used size"),
	)
	if err != nil {
		return nil, err
	}

	TotalSize, err := m.Int64ObservableGauge(
		"telemetry.stream.total_size",
		metric.WithUnit("By"),
		metric.WithDescription("Total size"),
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
	_, err := m.m.RegisterCallback(cb, m.UsedSize, m.TotalSize)
	return err
}
