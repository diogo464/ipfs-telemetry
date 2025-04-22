package telemetry

import (
	"github.com/diogo464/telemetry/internal/stream"
)

type serviceMetrics struct {
	stream *stream.Stream
}

func newServiceMetrics(stream *stream.Stream) *serviceMetrics {
	return &serviceMetrics{
		stream: stream,
	}
}
