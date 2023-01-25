package otlp_exporter

import (
	"context"

	"github.com/diogo464/telemetry/internal/otlp_exporter/transform"
	"github.com/diogo464/telemetry/internal/stream"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/protobuf/proto"
)

var _ sdk_metric.Exporter = (*exporter)(nil)

type exporter struct {
	stream              *stream.Stream
	aggregationSelector sdk_metric.AggregationSelector
	temporalitySelector sdk_metric.TemporalitySelector
}

func New(stream *stream.Stream) sdk_metric.Exporter {
	return &exporter{
		stream:              stream,
		aggregationSelector: sdk_metric.DefaultAggregationSelector,
		temporalitySelector: sdk_metric.DefaultTemporalitySelector,
	}
}

// Aggregation implements metric.Exporter
func (e *exporter) Aggregation(kind sdk_metric.InstrumentKind) aggregation.Aggregation {
	return e.aggregationSelector(kind)
}

// Export implements metric.Exporter
func (e *exporter) Export(ctx context.Context, rm metricdata.ResourceMetrics) error {
	pbrm, err := transform.ResourceMetrics(rm)
	if err != nil {
		return err
	}
	data, err := proto.Marshal(pbrm)
	if err != nil {
		return err
	}
	return e.stream.Write(data)
}

// ForceFlush implements metric.Exporter
func (*exporter) ForceFlush(ctx context.Context) error {
	return nil
}

// Shutdown implements metric.Exporter
func (*exporter) Shutdown(context.Context) error {
	return nil
}

// Temporality implements metric.Exporter
func (e *exporter) Temporality(kind sdk_metric.InstrumentKind) metricdata.Temporality {
	return e.temporalitySelector(kind)
}
