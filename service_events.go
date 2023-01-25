package telemetry

import (
	"sync"

	"github.com/diogo464/telemetry/internal/pb"
)

type eventId struct {
	scope       string
	name        string
	description string
}

func newEventIdFromConfig(config eventConfig) eventId {
	return eventId{
		scope:       config.Scope,
		name:        config.Name,
		description: config.Description,
	}
}

type serviceEvent struct {
	config  eventConfig
	stream  *serviceStream
	emitter *eventEmitter
}

type serviceEvents struct {
	streams *serviceStreams

	mu          sync.Mutex
	events      map[eventId]*serviceEvent
	descriptors []*pb.EventDescriptor
}

func newServiceEvents(streams *serviceStreams) *serviceEvents {
	return &serviceEvents{
		streams: streams,

		events:      make(map[eventId]*serviceEvent),
		descriptors: make([]*pb.EventDescriptor, 0),
	}
}

func (e *serviceEvents) copyDescriptors() []*pb.EventDescriptor {
	e.mu.Lock()
	defer e.mu.Unlock()

	descriptors := make([]*pb.EventDescriptor, len(e.descriptors))
	copy(descriptors, e.descriptors)

	return descriptors
}

func (e *serviceEvents) create(config eventConfig) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := newEventIdFromConfig(config)
	if se, ok := e.events[id]; ok {
		return se.emitter
	}

	stream := e.streams.create()
	emitter := newEventEmitter(config, stream.stream)

	e.events[id] = &serviceEvent{
		config:  config,
		stream:  stream,
		emitter: emitter,
	}

	e.descriptors = append(e.descriptors, eventDescriptorToPb(EventDescriptor{
		StreamId:    stream.streamId,
		Scope:       config.Scope,
		Name:        config.Name,
		Description: config.Description,
	}))

	return emitter
}
