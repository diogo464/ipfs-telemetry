package telemetry

import (
	"encoding/json"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

var _ (EventEmitter) = (*eventEmitter)(nil)
var _ (EventEmitter) = (*noOpEventEmitter)(nil)

type EventEmitter interface {
	Emit(interface{})
}

type EventDescriptor struct {
	Scope       instrumentation.Scope `json:"scope"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
}

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

type eventEmitter struct {
	name   string
	stream *stream.Stream
}

func newEventEmitter(streams *serviceStreams, desc EventDescriptor) *eventEmitter {
	streamType := &pb.StreamType{
		Type: &pb.StreamType_Event{
			Event: eventDescriptorToPb(desc),
		},
	}
	sstream := streams.create(streamType)
	return &eventEmitter{
		name:   desc.Name,
		stream: sstream.stream,
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
		Scope: &v1.InstrumentationScope{
			Name:    descriptor.Scope.Name,
			Version: descriptor.Scope.Version,
		},
		Name:        descriptor.Name,
		Description: descriptor.Description,
	}
}
