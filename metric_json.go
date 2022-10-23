package telemetry

import (
	"context"
	"encoding/json"
)

var _ (Collector) = (*JsonMetric)(nil)

type JsonMetric struct {
	descriptor CollectorDescriptor
	callback   func(context.Context) (interface{}, error)
}

func NewJsonMetric(name string, cb func(context.Context) (interface{}, error), opts ...CollectorOption) *JsonMetric {
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
	return &JsonMetric{
		descriptor: descriptor,
		callback:   cb,
	}
}

// Close implements Collector
func (*JsonMetric) Close() {
}

// Collect implements Collector
func (m *JsonMetric) Collect(ctx context.Context, s *Stream) error {
	value, err := m.callback(ctx)
	if err != nil {
		return err
	}
	marshaled, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.Write(marshaled)
}

// Descriptor implements Collector
func (m *JsonMetric) Descriptor() CollectorDescriptor {
	return m.descriptor
}

// Open implements Collector
func (*JsonMetric) Open() {
}
