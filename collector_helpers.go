package telemetry

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/atomic"
)

var _ (Collector) = (*simpleSnapshot)(nil)
var _ (Collector) = (*jsonSnapshot)(nil)

var (
	collectorDefaultPeriod = time.Second * 15
)

type EventCollector[T any] interface {
	Collector
	Publish(*T) error
}

type simpleSnapshot struct {
	name      string
	encoding  string
	period    time.Duration
	collectFn func(*Stream, context.Context) error
}

// Close implements Collector
func (*simpleSnapshot) Close() {
}

// Collect implements Collector
func (b *simpleSnapshot) Collect(ctx context.Context, s *Stream) error {
	return b.collectFn(s, ctx)
}

// Descriptor implements Collector
func (b *simpleSnapshot) Descriptor() CollectorDescriptor {
	return CollectorDescriptor{
		Name:     b.name,
		Period:   b.period,
		Encoding: b.encoding,
	}
}

// Open implements Collector
func (*simpleSnapshot) Open() {
}

type SnapshotDescriptor struct {
	Name      string
	Encoding  string
	Period    time.Duration
	CollectFn func(*Stream, context.Context) error
}

func SimpleSnapshot(d SnapshotDescriptor) Collector {
	if d.Period == 0 {
		d.Period = collectorDefaultPeriod
	}
	if d.Encoding == "" {
		d.Encoding = ENCODING_UNKNOWN
	}

	return &simpleSnapshot{
		name:      d.Name,
		period:    d.Period,
		collectFn: d.CollectFn,
	}
}

type jsonSnapshot struct {
	name      string
	period    time.Duration
	collectFn func(context.Context) (interface{}, error)
}

// Open implements Collector
func (*jsonSnapshot) Open() {
}

// Close implements Collector
func (*jsonSnapshot) Close() {
}

// Collect implements Collector
func (s *jsonSnapshot) Collect(ctx context.Context, stream *Stream) error {
	data, err := s.collectFn(ctx)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return stream.Write(encoded)
}

// Descriptor implements Collector
func (s *jsonSnapshot) Descriptor() CollectorDescriptor {
	return CollectorDescriptor{
		Name:     s.name,
		Period:   s.period,
		Encoding: ENCODING_JSON,
	}
}

type JsonSnapshotDescriptor struct {
	Name      string
	Period    time.Duration
	CollectFn func(context.Context) (interface{}, error)
}

func JsonSnapshot(d JsonSnapshotDescriptor) Collector {
	if d.Period == 0 {
		d.Period = collectorDefaultPeriod
	}

	return &jsonSnapshot{
		name:      d.Name,
		period:    d.Period,
		collectFn: d.CollectFn,
	}
}

type jsonEvent[T any] struct {
	name    string
	period  time.Duration
	open    *atomic.Bool
	mu      sync.Mutex
	buffers [][]byte
}

// Open implements EventCollector
func (e *jsonEvent[T]) Open() {
	e.open.Store(true)
}

// Close implements EventCollector
func (e *jsonEvent[T]) Close() {
	e.open.Store(false)
}

// Collect implements EventCollector
func (e *jsonEvent[T]) Collect(_ context.Context, stream *Stream) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, buf := range e.buffers {
		if err := stream.Write(buf); err != nil {
			return err
		}
	}
	e.buffers = nil
	return nil
}

// Descriptor implements EventCollector
func (e *jsonEvent[T]) Descriptor() CollectorDescriptor {
	return CollectorDescriptor{
		Name:     e.name,
		Period:   e.period,
		Encoding: ENCODING_JSON,
	}
}

// Publish implements EventCollector
func (e *jsonEvent[T]) Publish(ev *T) error {
	if e.open.Load() {
		encoded, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		e.mu.Lock()
		defer e.mu.Unlock()
		e.buffers = append(e.buffers, encoded)
	}
	return nil
}

type JsonEventDescriptor struct {
	Name   string
	Period time.Duration
}

func JsonEvent[T any](e JsonEventDescriptor) EventCollector[T] {
	if e.Period == 0 {
		e.Period = collectorDefaultPeriod
	}
	return &jsonEvent[T]{
		name:   e.Name,
		period: e.Period,
		open:   atomic.NewBool(false),
	}
}
