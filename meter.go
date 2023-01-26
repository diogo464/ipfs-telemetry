package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
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

	af64 *serviceAsyncFloat64
	ai64 *serviceAsyncInt64
	sf64 *serviceSyncFloat64
	si64 *serviceSyncInt64
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

		af64: newServiceAsyncFloat64(service, name, meter.AsyncFloat64()),
		ai64: newServiceAsyncInt64(service, name, meter.AsyncInt64()),
		sf64: newServiceSyncFloat64(service, name, meter.SyncFloat64()),
		si64: newServiceSyncInt64(service, name, meter.SyncInt64()),
	}
}

// AsyncFloat64 implements Meter
func (m *serviceMeter) AsyncFloat64() asyncfloat64.InstrumentProvider {
	return m.af64
}

// AsyncInt64 implements Meter
func (m *serviceMeter) AsyncInt64() asyncint64.InstrumentProvider {
	return m.ai64
}

// RegisterCallback implements Meter
func (m *serviceMeter) RegisterCallback(insts []instrument.Asynchronous, function func(context.Context)) error {
	return m.meter.RegisterCallback(insts, function)
}

// SyncFloat64 implements Meter
func (m *serviceMeter) SyncFloat64() syncfloat64.InstrumentProvider {
	return m.sf64
}

// SyncInt64 implements Meter
func (m *serviceMeter) SyncInt64() syncint64.InstrumentProvider {
	return m.si64
}

// Property implements Meter
func (m *serviceMeter) Property(name string, value PropertyValue, opts ...instrument.Option) {
	config := instrument.NewConfig(opts...)
	m.service.properties.create(Property{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
		Value:       value,
	})
}

// Event implements Meter
func (m *serviceMeter) Event(name string, opts ...instrument.Option) EventEmitter {
	config := instrument.NewConfig(opts...)
	return m.service.events.create(EventDescriptor{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
	})
}

// PeriodicEvent implements Meter
func (e *serviceMeter) PeriodicEvent(ctx context.Context, name string, interval time.Duration, cb func(context.Context, EventEmitter) error, opts ...instrument.Option) {
	config := instrument.NewConfig(opts...)
	e.service.events.createPeriodic(EventDescriptor{
		Scope:       e.scope,
		Name:        name,
		Description: config.Description(),
	}, ctx, interval, cb)
}
