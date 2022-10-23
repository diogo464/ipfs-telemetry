package telemetry

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/diogo464/telemetry/vecdeque"
)

var _ (Collector) = (*BackloggedJsonMetric[any])(nil)

type backloggedJsonMetricItem[T any] struct {
	timestamp uint64
	value     *T
}

type BackloggedJsonMetric[T any] struct {
	descriptor CollectorDescriptor
	backlog_mu *sync.Mutex
	backlog    *vecdeque.VecDeque[backloggedJsonMetricItem[T]]
}

func NewBackloggedJsonMetric[T any](name string, opts ...CollectorOption) *BackloggedJsonMetric[T] {
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
	return &BackloggedJsonMetric[T]{
		descriptor: descriptor,
		backlog_mu: &sync.Mutex{},
		backlog:    vecdeque.New[backloggedJsonMetricItem[T]](),
	}
}

func (m *BackloggedJsonMetric[T]) Push(value *T) {
	m.backlog_mu.Lock()
	defer m.backlog_mu.Unlock()
	m.backlog.PushBack(backloggedJsonMetricItem[T]{
		timestamp: TimestampNow(),
		value:     value,
	})
}

// Close implements Collector
func (*BackloggedJsonMetric[T]) Close() {
}

// Collect implements Collector
func (m *BackloggedJsonMetric[T]) Collect(_ context.Context, s *Stream) error {
	m.backlog_mu.Lock()
	defer m.backlog_mu.Unlock()

	var err error = nil
	for !m.backlog.IsEmpty() {
		item := m.backlog.PopFront()
		marshaled, e := json.Marshal(item.value)
		if e != nil {
			err = e
			continue
		}
		if e := s.WriteWithTimestamp(item.timestamp, marshaled); e != nil {
			err = e
		}
	}

	return err
}

// Descriptor implements Collector
func (m *BackloggedJsonMetric[T]) Descriptor() CollectorDescriptor {
	return m.descriptor
}

// Open implements Collector
func (*BackloggedJsonMetric[T]) Open() {
}
