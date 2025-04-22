package telemetry

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var _ (MeterProvider) = (*noOpMeterProvider)(nil)

type noOpMeterProvider struct {
	noop.MeterProvider
}

func NewNoopMeterProvider() MeterProvider {
	return &noOpMeterProvider{MeterProvider: noop.NewMeterProvider()}
}

// Meter implements MeterProvider
func (n *noOpMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return n.TelemetryMeter(instrumentationName, opts...)
}

// TelemetryMeter implements MeterProvider
func (p *noOpMeterProvider) TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter {
	m := p.MeterProvider.Meter(instrumentationName, opts...)
	return &noOpMeter{m}
}
