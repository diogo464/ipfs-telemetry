package telemetry

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type eventId struct {
	scope       instrumentation.Scope
	name        string
	description string
}

func newEventIdFromDescriptor(desc EventDescriptor) eventId {
	return eventId{
		scope:       desc.Scope,
		name:        desc.Name,
		description: desc.Description,
	}
}

type serviceEvent struct {
	emitter *eventEmitter
}

type serviceEvents struct {
	streams *serviceStreams

	mu     sync.Mutex
	events map[eventId]*serviceEvent
}

func newServiceEvents(streams *serviceStreams) *serviceEvents {
	return &serviceEvents{
		streams: streams,

		events: make(map[eventId]*serviceEvent),
	}
}

func (e *serviceEvents) create(desc EventDescriptor) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := newEventIdFromDescriptor(desc)
	if se, ok := e.events[id]; ok {
		return se.emitter
	}

	emitter := newEventEmitter(e.streams, desc)
	e.events[id] = &serviceEvent{
		emitter: emitter,
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
