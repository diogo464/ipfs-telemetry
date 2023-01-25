package telemetry

import (
	"encoding/json"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
)

var _ (EventEmitter) = (*eventEmitter)(nil)
var _ (EventEmitter) = (*noOpEventEmitter)(nil)

type EventEmitter interface {
	Emit(interface{})
}

type EventDescriptor struct {
	StreamId    StreamId `json:"stream_id"`
	Scope       string   `json:"scope"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

type eventConfig struct {
	Scope       string
	Name        string
	Description string
}

type eventEmitter struct {
	name   string
	stream *stream.Stream
}

func newEventEmitter(config eventConfig, stream *stream.Stream) *eventEmitter {
	return &eventEmitter{
		name:   config.Name,
		stream: stream,
	}
}

// Emit implements EventEmitter
func (e *eventEmitter) Emit(data interface{}) {
	var s = e.stream
	if s != nil {
		if marshaled, err := json.Marshal(data); err == nil {
			s.Write(marshaled)
		} else {
			log.Warnf("failed to emit event",
				"event", e.name,
				"error", err)
		}
	}
}

type noOpEventEmitter struct {
}

// Emit implements EventEmitter
func (*noOpEventEmitter) Emit(interface{}) {
}

func eventDescriptorToPb(descriptor EventDescriptor) *pb.EventDescriptor {
	return &pb.EventDescriptor{
		StreamId:    uint32(descriptor.StreamId),
		Scope:       descriptor.Scope,
		Name:        descriptor.Name,
		Description: descriptor.Description,
	}
}

func eventDescriptorFromPb(descriptor *pb.EventDescriptor) EventDescriptor {
	return EventDescriptor{
		StreamId:    StreamId(descriptor.StreamId),
		Scope:       descriptor.Scope,
		Name:        descriptor.Name,
		Description: descriptor.Description,
	}
}
