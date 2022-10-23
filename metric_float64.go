package telemetry

import (
	"context"
	"encoding/binary"

	"go.uber.org/atomic"
)

var _ (Collector) = (*Float64Metric)(nil)

type Float64Metric struct {
	descriptor CollectorDescriptor
	counter    *atomic.Float64
}

func (m *Float64Metric) Inc() {
	m.counter.Add(1.0)
}

func (m *Float64Metric) Dec() {
	m.counter.Sub(1.0)
}

func (m *Float64Metric) Add(delta float64) {
	m.counter.Add(delta)
}

func (m *Float64Metric) Sub(delta float64) {
	m.counter.Sub(delta)
}

func (m *Float64Metric) Set(v float64) {
	m.counter.Store(v)
}

func NewFloat64Metric(name string, opts ...CollectorOption) *Float64Metric {
	config := collectorConfigDefaults()
	collectorConfigApply(config, opts...)
	descriptor := CollectorDescriptor{
		Name:     name,
		Period:   collectorDefaultPeriod,
		Encoding: EncodingFloat64,
	}
	if config.period != nil {
		descriptor.Period = *config.period
	}
	return &Float64Metric{
		descriptor: descriptor,
		counter:    atomic.NewFloat64(0),
	}
}

// Close implements Collector
func (*Float64Metric) Close() {
}

// Collect implements Collector
func (m *Float64Metric) Collect(_ context.Context, s *Stream) error {
	return s.AllocAndWrite(8, func(b []byte) error {
		binary.BigEndian.PutUint64(b, uint64(m.counter.Load()))
		return nil
	})
}

// Descriptor implements Collector
func (m *Float64Metric) Descriptor() CollectorDescriptor {
	return m.descriptor
}

// Open implements Collector
func (*Float64Metric) Open() {
}
