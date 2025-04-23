package telemetry

import (
	"sync"

	"github.com/diogo464/telemetry/internal/stream"
)

type serviceStreamStats struct {
	streamId StreamId
	stats    stream.Stats
}

type serviceStream struct {
	stream   *stream.Stream
	streamId StreamId
}

type serviceStreams struct {
	mu             sync.Mutex
	streams        map[StreamId]*serviceStream
	nextID         StreamId
	defaultOptions []stream.Option
}

func newServiceStreams(defaultOptions ...stream.Option) *serviceStreams {
	return &serviceStreams{
		streams:        make(map[StreamId]*serviceStream),
		defaultOptions: defaultOptions,
	}
}

func (s *serviceStreams) create(options ...stream.Option) *serviceStream {
	s.mu.Lock()
	defer s.mu.Unlock()

	options = append(options, s.defaultOptions...)

	stream := stream.New(options...)

	id := s.nextID
	s.nextID++

	s.streams[id] = &serviceStream{
		stream:   stream,
		streamId: id,
	}

	return s.streams[id]
}

func (s *serviceStreams) get(id StreamId) *serviceStream {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.streams[id]
}

func (s *serviceStreams) getSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.streams)
}

func (s *serviceStreams) getStats() []serviceStreamStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := make([]serviceStreamStats, 0, len(s.streams))
	for id, stream := range s.streams {
		stats = append(stats, serviceStreamStats{
			streamId: id,
			stats:    stream.stream.Stats(),
		})
	}

	return stats
}
