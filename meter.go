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
	scope   string
	config  metric.MeterConfig
	service *Service
	meter   metric.Meter
}

// AsyncFloat64 implements Meter
func (m *serviceMeter) AsyncFloat64() asyncfloat64.InstrumentProvider {
	provider := m.meter.AsyncFloat64()
	return &serviceAsyncFloat64{
		service:  m.service,
		scope:    m.scope,
		provider: provider,
	}
}

// AsyncInt64 implements Meter
func (m *serviceMeter) AsyncInt64() asyncint64.InstrumentProvider {
	provider := m.meter.AsyncInt64()
	return &serviceAsyncInt64{
		service:  m.service,
		scope:    m.scope,
		provider: provider,
	}
}

// RegisterCallback implements Meter
func (m *serviceMeter) RegisterCallback(insts []instrument.Asynchronous, function func(context.Context)) error {
	return m.meter.RegisterCallback(insts, function)
}

// SyncFloat64 implements Meter
func (m *serviceMeter) SyncFloat64() syncfloat64.InstrumentProvider {
	provider := m.meter.SyncFloat64()
	return &serviceSyncFloat64{
		service:  m.service,
		scope:    m.scope,
		provider: provider,
	}
}

// SyncInt64 implements Meter
func (m *serviceMeter) SyncInt64() syncint64.InstrumentProvider {
	provider := m.meter.SyncInt64()
	return &serviceSyncInt64{
		service:  m.service,
		scope:    m.scope,
		provider: provider,
	}
}

// Capture implements Meter
func (m *serviceMeter) Capture(name string, callback CaptureCallback, interval time.Duration, opts ...instrument.Option) {
	config := instrument.NewConfig(opts...)
	m.service.createCapture(captureConfig{
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
	return m.service.createEvent(eventConfig{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
	})
}

// Property implements Meter
func (m *serviceMeter) Property(name string, value PropertyValue, opts ...instrument.Option) {
	config := instrument.NewConfig(opts...)
	m.config.SchemaURL()
	m.service.createProperty(propertyConfig{
		Scope:       m.scope,
		Name:        name,
		Description: config.Description(),
		Value:       value,
	})
}
