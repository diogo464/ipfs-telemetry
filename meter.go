package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var _ Meter = (*serviceMeter)(nil)

type Meter interface {
	metric.Meter

	Property(name string, value PropertyValue, opts ...metric.InstrumentOption)

	Event(name string, opts ...metric.InstrumentOption) EventEmitter

	PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...metric.InstrumentOption)
}

type serviceMeter struct {
	metric.Meter
	service *Service
	scope   instrumentation.Scope
	config  metric.MeterConfig
}

func newServiceMeter(service *Service, name string, config metric.MeterConfig, meter metric.Meter) *serviceMeter {

	return &serviceMeter{
		Meter:   meter,
		service: service,
		scope: instrumentation.Scope{
			Name:      name,
			Version:   config.InstrumentationVersion(),
			SchemaURL: config.SchemaURL(),
		},
		config: config,
	}
}

// Property implements Meter
func (m *serviceMeter) Property(name string, value PropertyValue, opts ...metric.InstrumentOption) {
	desc, _ := decomposeInstrumentOptions(opts...)
	m.service.properties.create(Property{
		Scope:       m.scope,
		Name:        name,
		Description: desc,
		Value:       value,
	})
}

// Event implements Meter
func (m *serviceMeter) Event(name string, opts ...metric.InstrumentOption) EventEmitter {
	desc, _ := decomposeInstrumentOptions(opts...)
	return m.service.events.create(EventDescriptor{
		Scope:       m.scope,
		Name:        name,
		Description: desc,
	})
}

// PeriodicEvent implements Meter
func (e *serviceMeter) PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...metric.InstrumentOption) {
	desc, _ := decomposeInstrumentOptions(opts...)
	e.service.events.createPeriodic(EventDescriptor{
		Scope:       e.scope,
		Name:        name,
		Description: desc,
	}, ctx, interval, cb)
}

// Decompose metric.Options into a description and a unit.
func decomposeInstrumentOptions(opts ...metric.InstrumentOption) (string, string) {
	int64opts := make([]metric.Int64CounterOption, 0, len(opts))
	for _, opt := range opts {
		int64opts = append(int64opts, opt)
	}
	config := metric.NewInt64CounterConfig(int64opts...)
	return config.Description(), config.Unit()
}
