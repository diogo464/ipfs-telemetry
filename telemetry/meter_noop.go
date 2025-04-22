package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
)

var _ (Meter) = (*noOpMeter)(nil)

type noOpMeter struct {
	metric.Meter
}

// Property implements Meter
func (*noOpMeter) Property(name string, value PropertyValue, opts ...metric.InstrumentOption) {
}

// Event implements Meter
func (*noOpMeter) Event(name string, opts ...metric.InstrumentOption) EventEmitter {
	return &noOpEventEmitter{}
}

// PeriodicEvent implements Meter
func (*noOpMeter) PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...metric.InstrumentOption) {
}
