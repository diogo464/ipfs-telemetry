package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
)

var _ (Meter) = (*noOpMeter)(nil)

type noOpMeter struct {
	noop_meter metric.Meter
}

// Float64Counter implements Meter
func (m *noOpMeter) Float64Counter(name string, options ...instrument.Float64Option) (instrument.Float64Counter, error) {
	return m.noop_meter.Float64Counter(name, options...)
}

// Float64Histogram implements Meter
func (m *noOpMeter) Float64Histogram(name string, options ...instrument.Float64Option) (instrument.Float64Histogram, error) {
	return m.noop_meter.Float64Histogram(name, options...)
}

// Float64ObservableCounter implements Meter
func (m *noOpMeter) Float64ObservableCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableCounter, error) {
	return m.noop_meter.Float64ObservableCounter(name, options...)
}

// Float64ObservableGauge implements Meter
func (m *noOpMeter) Float64ObservableGauge(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableGauge, error) {
	return m.noop_meter.Float64ObservableGauge(name, options...)
}

// Float64ObservableUpDownCounter implements Meter
func (m *noOpMeter) Float64ObservableUpDownCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableUpDownCounter, error) {
	return m.noop_meter.Float64ObservableUpDownCounter(name, options...)
}

// Float64UpDownCounter implements Meter
func (m *noOpMeter) Float64UpDownCounter(name string, options ...instrument.Float64Option) (instrument.Float64UpDownCounter, error) {
	return m.noop_meter.Float64UpDownCounter(name, options...)
}

// Int64Counter implements Meter
func (m *noOpMeter) Int64Counter(name string, options ...instrument.Int64Option) (instrument.Int64Counter, error) {
	return m.noop_meter.Int64Counter(name, options...)
}

// Int64Histogram implements Meter
func (m *noOpMeter) Int64Histogram(name string, options ...instrument.Int64Option) (instrument.Int64Histogram, error) {
	return m.noop_meter.Int64Histogram(name, options...)
}

// Int64ObservableCounter implements Meter
func (m *noOpMeter) Int64ObservableCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableCounter, error) {
	return m.noop_meter.Int64ObservableCounter(name, options...)
}

// Int64ObservableGauge implements Meter
func (m *noOpMeter) Int64ObservableGauge(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableGauge, error) {
	return m.noop_meter.Int64ObservableGauge(name, options...)
}

// Int64ObservableUpDownCounter implements Meter
func (m *noOpMeter) Int64ObservableUpDownCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableUpDownCounter, error) {
	return m.noop_meter.Int64ObservableUpDownCounter(name, options...)
}

// Int64UpDownCounter implements Meter
func (m *noOpMeter) Int64UpDownCounter(name string, options ...instrument.Int64Option) (instrument.Int64UpDownCounter, error) {
	return m.noop_meter.Int64UpDownCounter(name, options...)
}

// RegisterCallback implements Meter
func (m *noOpMeter) RegisterCallback(f metric.Callback, instruments ...instrument.Asynchronous) (metric.Registration, error) {
	return m.noop_meter.RegisterCallback(f, instruments...)
}

// Property implements Meter
func (*noOpMeter) Property(name string, value PropertyValue, opts ...instrument.Option) {
}

// Event implements Meter
func (*noOpMeter) Event(name string, opts ...instrument.Option) EventEmitter {
	return &noOpEventEmitter{}
}

// PeriodicEvent implements Meter
func (*noOpMeter) PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...instrument.Option) {
}
