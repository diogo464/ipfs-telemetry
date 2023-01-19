package telemetry

import (
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

var _ (asyncint64.InstrumentProvider) = (*serviceAsyncInt64)(nil)
var _ (asyncfloat64.InstrumentProvider) = (*serviceAsyncFloat64)(nil)
var _ (syncint64.InstrumentProvider) = (*serviceSyncInt64)(nil)
var _ (syncfloat64.InstrumentProvider) = (*serviceSyncFloat64)(nil)

type serviceAsyncInt64 struct {
	service  *Service
	scope    string
	provider asyncint64.InstrumentProvider
}

// Counter implements asyncint64.InstrumentProvider
func (s *serviceAsyncInt64) Counter(name string, opts ...instrument.Option) (asyncint64.Counter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Counter(name, opts...)
}

// Gauge implements asyncint64.InstrumentProvider
func (s *serviceAsyncInt64) Gauge(name string, opts ...instrument.Option) (asyncint64.Gauge, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Gauge(name, opts...)
}

// UpDownCounter implements asyncint64.InstrumentProvider
func (s *serviceAsyncInt64) UpDownCounter(name string, opts ...instrument.Option) (asyncint64.UpDownCounter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.UpDownCounter(name, opts...)
}

type serviceAsyncFloat64 struct {
	service  *Service
	scope    string
	provider asyncfloat64.InstrumentProvider
}

// Counter implements asyncfloat64.InstrumentProvider
func (s *serviceAsyncFloat64) Counter(name string, opts ...instrument.Option) (asyncfloat64.Counter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Counter(name, opts...)
}

// Gauge implements asyncfloat64.InstrumentProvider
func (s *serviceAsyncFloat64) Gauge(name string, opts ...instrument.Option) (asyncfloat64.Gauge, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Gauge(name, opts...)
}

// UpDownCounter implements asyncfloat64.InstrumentProvider
func (s *serviceAsyncFloat64) UpDownCounter(name string, opts ...instrument.Option) (asyncfloat64.UpDownCounter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.UpDownCounter(name, opts...)
}

type serviceSyncInt64 struct {
	service  *Service
	scope    string
	provider syncint64.InstrumentProvider
}

// Counter implements syncint64.InstrumentProvider
func (s *serviceSyncInt64) Counter(name string, opts ...instrument.Option) (syncint64.Counter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Counter(name, opts...)
}

// Histogram implements syncint64.InstrumentProvider
func (s *serviceSyncInt64) Histogram(name string, opts ...instrument.Option) (syncint64.Histogram, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Histogram(name, opts...)
}

// UpDownCounter implements syncint64.InstrumentProvider
func (s *serviceSyncInt64) UpDownCounter(name string, opts ...instrument.Option) (syncint64.UpDownCounter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.UpDownCounter(name, opts...)
}

type serviceSyncFloat64 struct {
	service  *Service
	scope    string
	provider syncfloat64.InstrumentProvider
}

// Counter implements syncfloat64.InstrumentProvider
func (s *serviceSyncFloat64) Counter(name string, opts ...instrument.Option) (syncfloat64.Counter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Counter(name, opts...)
}

// Histogram implements syncfloat64.InstrumentProvider
func (s *serviceSyncFloat64) Histogram(name string, opts ...instrument.Option) (syncfloat64.Histogram, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.Histogram(name, opts...)
}

// UpDownCounter implements syncfloat64.InstrumentProvider
func (s *serviceSyncFloat64) UpDownCounter(name string, opts ...instrument.Option) (syncfloat64.UpDownCounter, error) {
	cfg := instrument.NewConfig(opts...)
	desc := MetricDescriptor{Scope: s.scope, Name: name, Description: cfg.Description()}
	s.service.addMetricDescriptor(desc)
	return s.provider.UpDownCounter(name, opts...)
}
