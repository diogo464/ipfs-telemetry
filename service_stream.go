package telemetry

import (
	"sync"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
)

type serviceStream struct {
	stream           *stream.Stream
	streamId         StreamId
	streamDescriptor *pb.StreamDescriptor
}

type serviceStreams struct {
	mu             sync.Mutex
	streams        map[StreamId]*serviceStream
	descriptors    []*pb.StreamDescriptor
	nextID         StreamId
	defaultOptions []stream.Option
}

func newServiceStreams(defaultOptions ...stream.Option) *serviceStreams {
	return &serviceStreams{
		streams:        make(map[StreamId]*serviceStream),
		defaultOptions: defaultOptions,
	}
}

func (s *serviceStreams) copyDescriptors() []*pb.StreamDescriptor {
	s.mu.Lock()
	defer s.mu.Unlock()

	descriptors := make([]*pb.StreamDescriptor, len(s.descriptors))
	copy(descriptors, s.descriptors)

	return descriptors
}

func (s *serviceStreams) create(ty *pb.StreamType, options ...stream.Option) *serviceStream {
	s.mu.Lock()
	defer s.mu.Unlock()

	options = append(options, s.defaultOptions...)

	stream := stream.New(options...)

	id := s.nextID
	s.nextID++

	s.streams[id] = &serviceStream{
		stream:   stream,
		streamId: id,
		streamDescriptor: &pb.StreamDescriptor{
			StreamId:   uint32(id),
			StreamType: ty,
		},
	}
	s.descriptors = append(s.descriptors, s.streams[id].streamDescriptor)

	return s.streams[id]
}

func (s *serviceStreams) has(id StreamId) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.streams[id]
	return ok
}

func (s *serviceStreams) get(id StreamId) *serviceStream {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.streams[id]
}
