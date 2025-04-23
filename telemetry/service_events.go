package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

type eventId uint32

type serviceEvent struct {
	emitter    *eventEmitter
	descriptor *pb.EventDescriptor
}

type serviceEvents struct {
	streams *serviceStreams

	mu     sync.Mutex
	events map[eventId]*serviceEvent
	nextId eventId
}

func newServiceEvents(streams *serviceStreams) *serviceEvents {
	return &serviceEvents{
		streams: streams,

		events: make(map[eventId]*serviceEvent),
		nextId: 0,
	}
}

func (e *serviceEvents) create(desc EventDescriptor) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := e.nextId
	e.nextId += 1

	if se, ok := e.events[id]; ok {
		return se.emitter
	}

	stream := e.streams.create()
	emitter := newEventEmitter(stream.stream, desc)
	e.events[id] = &serviceEvent{
		emitter: emitter,
		descriptor: &pb.EventDescriptor{
			EventId: uint32(id),
			Scope: &v1.InstrumentationScope{
				Name:    desc.Scope.Name,
				Version: desc.Scope.Version,
			},
			Name:        desc.Name,
			Description: desc.Description,
		},
	}

	return emitter
}

func (e *serviceEvents) createPeriodic(desc EventDescriptor, ctx context.Context, interval time.Duration, cb func(context.Context, EventEmitter) error) {
	emitter := e.create(desc)
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := cb(ctx, emitter); err != nil {
					log.Warnf("error while emitting periodic event: %v", err)
				}
			}
		}
	}()
}

func (e *serviceEvents) getSize() int {
	e.mu.Lock()
	defer e.mu.Unlock()

	return len(e.events)
}

func (e *serviceEvents) getEventStreamById(id eventId) *stream.Stream {
	if event, ok := e.events[id]; ok {
		return event.emitter.stream
	} else {
		return nil
	}
}

func (e *serviceEvents) getEventDescriptors() []*pb.EventDescriptor {
	descriptors := make([]*pb.EventDescriptor, 0, len(e.events))
	for _, e := range e.events {
		descriptors = append(descriptors, e.descriptor)
	}
	return descriptors
}
