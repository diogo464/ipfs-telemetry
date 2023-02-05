package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var _ Meter = (*serviceMeter)(nil)

type Meter interface {
	metric.Meter

	Property(name string, value PropertyValue, opts ...instrument.Option)

	Event(name string, opts ...instrument.Option) EventEmitter

	PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...instrument.Option)
}

type serviceMeter struct {
	service *Service

	scope  instrumentation.Scope
	config metric.MeterConfig
	meter  metric.Meter
}

// Float64Counter implements Meter
func (m *serviceMeter) Float64Counter(name string, options ...instrument.Float64Option) (instrument.Float64Counter, error) {
	return m.meter.Float64Counter(name, options...)
}

// Float64Histogram implements Meter
func (m *serviceMeter) Float64Histogram(name string, options ...instrument.Float64Option) (instrument.Float64Histogram, error) {
	return m.meter.Float64Histogram(name, options...)
}

// Float64ObservableCounter implements Meter
func (m *serviceMeter) Float64ObservableCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableCounter, error) {
	return m.meter.Float64ObservableCounter(name, options...)
}

// Float64ObservableGauge implements Meter
func (m *serviceMeter) Float64ObservableGauge(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableGauge, error) {
	return m.meter.Float64ObservableGauge(name, options...)
}

// Float64ObservableUpDownCounter implements Meter
func (m *serviceMeter) Float64ObservableUpDownCounter(name string, options ...instrument.Float64ObserverOption) (instrument.Float64ObservableUpDownCounter, error) {
	return m.meter.Float64ObservableUpDownCounter(name, options...)
}

// Float64UpDownCounter implements Meter
func (m *serviceMeter) Float64UpDownCounter(name string, options ...instrument.Float64Option) (instrument.Float64UpDownCounter, error) {
	return m.meter.Float64UpDownCounter(name, options...)
}

// Int64Counter implements Meter
func (m *serviceMeter) Int64Counter(name string, options ...instrument.Int64Option) (instrument.Int64Counter, error) {
	return m.meter.Int64Counter(name, options...)
}

// Int64Histogram implements Meter
func (m *serviceMeter) Int64Histogram(name string, options ...instrument.Int64Option) (instrument.Int64Histogram, error) {
	return m.meter.Int64Histogram(name, options...)
}

// Int64ObservableCounter implements Meter
func (m *serviceMeter) Int64ObservableCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableCounter, error) {
	return m.meter.Int64ObservableCounter(name, options...)
}

// Int64ObservableGauge implements Meter
func (m *serviceMeter) Int64ObservableGauge(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableGauge, error) {
	return m.meter.Int64ObservableGauge(name, options...)
}

// Int64ObservableUpDownCounter implements Meter
func (m *serviceMeter) Int64ObservableUpDownCounter(name string, options ...instrument.Int64ObserverOption) (instrument.Int64ObservableUpDownCounter, error) {
	return m.meter.Int64ObservableUpDownCounter(name, options...)
}

// Int64UpDownCounter implements Meter
func (m *serviceMeter) Int64UpDownCounter(name string, options ...instrument.Int64Option) (instrument.Int64UpDownCounter, error) {
	return m.meter.Int64UpDownCounter(name, options...)
}

// RegisterCallback implements Meter
func (m *serviceMeter) RegisterCallback(f metric.Callback, instruments ...instrument.Asynchronous) (metric.Registration, error) {
	return m.meter.RegisterCallback(f, instruments...)
}

func newServiceMeter(service *Service, name string, config metric.MeterConfig, meter metric.Meter) *serviceMeter {

	return &serviceMeter{
		service: service,

		scope: instrumentation.Scope{
			Name:      name,
			Version:   config.InstrumentationVersion(),
			SchemaURL: config.SchemaURL(),
		},
		config: config,
		meter:  meter,
	}
}

// Property implements Meter
func (m *serviceMeter) Property(name string, value PropertyValue, opts ...instrument.Option) {
	desc, _ := decomposeInstrumentOptions(opts...)
	m.service.properties.create(Property{
		Scope:       m.scope,
		Name:        name,
		Description: desc,
		Value:       value,
	})
}

// Event implements Meter
func (m *serviceMeter) Event(name string, opts ...instrument.Option) EventEmitter {
	desc, _ := decomposeInstrumentOptions(opts...)
	return m.service.events.create(EventDescriptor{
		Scope:       m.scope,
		Name:        name,
		Description: desc,
	})
}

// PeriodicEvent implements Meter
func (e *serviceMeter) PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...instrument.Option) {
	desc, _ := decomposeInstrumentOptions(opts...)
	e.service.events.createPeriodic(EventDescriptor{
		Scope:       e.scope,
		Name:        name,
		Description: desc,
	}, ctx, interval, cb)
}

// Decompose instrument.Options into a description and a unit.
func decomposeInstrumentOptions(opts ...instrument.Option) (string, unit.Unit) {
	int64opts := make([]instrument.Int64Option, 0, len(opts))
	for _, opt := range opts {
		int64opts = append(int64opts, opt)
	}
	config := instrument.NewInt64Config(int64opts...)
	return config.Description(), config.Unit()
}
