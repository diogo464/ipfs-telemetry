package telemetry

import (
	"encoding/json"
	"fmt"
)

var ErrEventAlreadyRegistered = fmt.Errorf("event already registered")

var _ (EventCollector) = (*Event)(nil)

type EventCollector interface {
	Open(*Stream)
	Descriptor() EventDescriptor
	Close()
}

type EventDescriptor struct {
	Name string
}

type EventConfig struct {
	Name string
}

type Event struct {
	descriptor EventDescriptor
	stream     *Stream
}

func NewEvent(config EventConfig) *Event {
	return &Event{
		descriptor: EventDescriptor{
			Name: config.Name,
		},
		stream: nil,
	}
}

func (e *Event) Emit(data interface{}) {
	var s = e.stream
	if s != nil {
		if marshaled, err := json.Marshal(data); err == nil {
			s.Write(marshaled)
		} else {
			log.Warnf("Failed to emit event",
				"event", e.descriptor.Name,
				"error", err)
		}
	}
}

// Open implements EventCollector
func (e *Event) Open(s *Stream) {
	if e.stream == nil {
		e.stream = s
	}
}

// Descriptor implements EventCollector
func (e *Event) Descriptor() EventDescriptor {
	return e.descriptor
}

// Close implements EventCollector
func (e *Event) Close() {
	e.stream = nil
}
