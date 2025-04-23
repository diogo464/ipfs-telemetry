package telemetry

import (
	"github.com/diogo464/telemetry/internal/stream"
)

type serviceMetrics struct {
	streamId StreamId
	stream   *stream.Stream
}

func newServiceMetrics(streams *serviceStreams) *serviceMetrics {
	metricsStream := streams.create()
	return &serviceMetrics{
		streamId: metricsStream.streamId,
		stream:   metricsStream.stream,
	}
}
