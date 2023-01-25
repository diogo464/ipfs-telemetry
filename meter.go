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
)

var _ Meter = (*serviceMeter)(nil)

type Meter interface {
	metric.Meter

	Property(name string, value PropertyValue, opts ...instrument.Option)

	Capture(name string, callback CaptureCallback, interval time.Duration, opts ...instrument.Option)

	Event(name string, opts ...instrument.Option) EventEmitter
}

type serviceMeter struct {
	service *Service

	scope  string
	config metric.MeterConfig
	meter  metric.Meter

	af64 *serviceAsyncFloat64
	ai64 *serviceAsyncInt64
	sf64 *serviceSyncFloat64
	si64 *serviceSyncInt64
}

func newServiceMeter(service *Service, scope string, config metric.MeterConfig, meter metric.Meter) *serviceMeter {
	return &serviceMeter{
		service: service,

		scope:  scope,
		config: config,
		meter:  meter,

		af64: newServiceAsyncFloat64(service, scope, meter.AsyncFloat64()),
		ai64: newServiceAsyncInt64(service, scope, meter.AsyncInt64()),
		sf64: newServiceSyncFloat64(service, scope, meter.SyncFloat64()),
		si64: newServiceSyncInt64(service, scope, meter.SyncInt64()),
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

// Capture implements Meter
func (m *serviceMeter) Capture(name string, callback CaptureCallback, interval time.Duration, opts ...instrument.Option) {
	config := instrument.NewConfig(opts...)
	m.service.captures.create(captureConfig{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
		Callback:    callback,
		Interval:    interval,
	})
}

// Event implements Meter
func (m *serviceMeter) Event(name string, opts ...instrument.Option) EventEmitter {
	config := instrument.NewConfig(opts...)
	return m.service.events.create(eventConfig{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
	})
}

// Property implements Meter
func (m *serviceMeter) Property(name string, value PropertyValue, opts ...instrument.Option) {
	config := instrument.NewConfig(opts...)
	m.service.properties.create(propertyConfig{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
		Value:       value,
	})
}
