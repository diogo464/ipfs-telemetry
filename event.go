package telemetry

import (
	"encoding/json"
)

var _ (EventEmitter) = (*eventEmitter)(nil)

type eventEmitter struct {
	name   string
	stream *Stream
}

func newEventEmitter(config EventConfig, stream *Stream) *eventEmitter {
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
			log.Warnf("Failed to emit event",
				"event", e.name,
				"error", err)
		}
	}
}
