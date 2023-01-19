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

var _ (Meter) = (*noOpMeter)(nil)

type noOpMeter struct {
	noop_meter metric.Meter
}

// AsyncFloat64 implements Meter
func (m *noOpMeter) AsyncFloat64() asyncfloat64.InstrumentProvider {
	return m.noop_meter.AsyncFloat64()
}

// AsyncInt64 implements Meter
func (m *noOpMeter) AsyncInt64() asyncint64.InstrumentProvider {
	return m.noop_meter.AsyncInt64()
}

// RegisterCallback implements Meter
func (m *noOpMeter) RegisterCallback(insts []instrument.Asynchronous, function func(context.Context)) error {
	return m.noop_meter.RegisterCallback(insts, function)
}

// SyncFloat64 implements Meter
func (m *noOpMeter) SyncFloat64() syncfloat64.InstrumentProvider {
	return m.noop_meter.SyncFloat64()
}

// SyncInt64 implements Meter
func (m *noOpMeter) SyncInt64() syncint64.InstrumentProvider {
	return m.noop_meter.SyncInt64()
}

// Capture implements Meter
func (*noOpMeter) Capture(name string, callback CaptureCallback, interval time.Duration, opts ...instrument.Option) {
}

// Event implements Meter
func (*noOpMeter) Event(name string, opts ...instrument.Option) EventEmitter {
	return &noOpEventEmitter{}
}

// Property implements Meter
func (*noOpMeter) Property(name string, value PropertyValue, opts ...instrument.Option) {
}
