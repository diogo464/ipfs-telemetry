package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"
)

// Meter implements Telemetry
func (s *Service) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return s.meter_provider.Meter(instrumentationName, opts...)
}

// UploadMetrics implements otlpmetric.Client
func (s *Service) UploadMetrics(_ context.Context, m *mpb.ResourceMetrics) error {
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	fmt.Println("Writting ", len(data), " bytes of metrics to the stream")
	return s.metrics.stream.Write(data)
}

// ForceFlush implements otlpmetric.Client
func (*Service) ForceFlush(context.Context) error {
	return nil
}

// Shutdown implements otlpmetric.Client
func (*Service) Shutdown(context.Context) error {
	return nil
}
