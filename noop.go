package telemetry

import "go.opentelemetry.io/otel/metric"

var _ (EventEmitter) = (*noOpEventEmitter)(nil)
var _ (Telemetry) = (*noOpTelemetry)(nil)

type noOpEventEmitter struct {
}

// Emit implements EventEmitter
func (*noOpEventEmitter) Emit(interface{}) {
}

type noOpTelemetry struct {
	provider metric.MeterProvider
}

func NewNoOpTelemetry() Telemetry {
	return &noOpTelemetry{
		provider: metric.NewNoopMeterProvider(),
	}
}

// Meter implements Telemetry
func (t *noOpTelemetry) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return t.provider.Meter(instrumentationName, opts...)
}

// Capture implements Telemetry
func (*noOpTelemetry) Capture(CaptureConfig) {
}

// Property implements Telemetry
func (*noOpTelemetry) Property(PropertyConfig) {
}

// Event implements Telemetry
func (*noOpTelemetry) Event(EventConfig) EventEmitter {
	return &noOpEventEmitter{}
}
