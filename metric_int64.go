package telemetry

import (
	"context"
	"encoding/binary"

	"go.uber.org/atomic"
)

var _ (Collector) = (*Int64Metric)(nil)

type Int64Metric struct {
	descriptor CollectorDescriptor
	counter    *atomic.Int64
}

func (m *Int64Metric) Inc() {
	m.counter.Inc()
}

func (m *Int64Metric) Dec() {
	m.counter.Dec()
}

func (m *Int64Metric) Add(delta int64) {
	m.counter.Add(delta)
}

func (m *Int64Metric) Sub(delta int64) {
	m.counter.Sub(delta)
}

func (m *Int64Metric) Set(v int64) {
	m.counter.Store(v)
}

func NewInt64Metric(name string, opts ...CollectorOption) *Int64Metric {
	config := collectorConfigDefaults()
	collectorConfigApply(config, opts...)
	descriptor := CollectorDescriptor{
		Name:     name,
		Period:   collectorDefaultPeriod,
		Encoding: EncodingInt64,
	}
	if config.period != nil {
		descriptor.Period = *config.period
	}
	return &Int64Metric{
		descriptor: descriptor,
		counter:    atomic.NewInt64(0),
	}
}

// Close implements Collector
func (*Int64Metric) Close() {
}

// Collect implements Collector
func (m *Int64Metric) Collect(_ context.Context, s *Stream) error {
	return s.AllocAndWrite(8, func(b []byte) error {
		binary.BigEndian.PutUint64(b, uint64(m.counter.Load()))
		return nil
	})
}

// Descriptor implements Collector
func (m *Int64Metric) Descriptor() CollectorDescriptor {
	return m.descriptor
}

// Open implements Collector
func (*Int64Metric) Open() {
}
