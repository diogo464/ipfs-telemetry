package telemetry

import "go.opentelemetry.io/otel/metric"

var _ (MeterProvider) = (*noOpMeterProvider)(nil)

type noOpMeterProvider struct {
}

func NewNoopMeterProvider() MeterProvider {
	return &noOpMeterProvider{}
}

// Meter implements MeterProvider
func (n *noOpMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return n.TelemetryMeter(instrumentationName, opts...)
}

// TelemetryMeter implements MeterProvider
func (*noOpMeterProvider) TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter {
	n := metric.NewNoopMeterProvider()
	m := n.Meter(instrumentationName, opts...)
	return &noOpMeter{noop_meter: m}
}
