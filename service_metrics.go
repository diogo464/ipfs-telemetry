package telemetry

import (
	"sync"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
)

type serviceMetrics struct {
	mu          sync.Mutex
	stream      *stream.Stream
	descriptors []*pb.MetricDescriptor
}

func newServiceMetrics(stream *stream.Stream) *serviceMetrics {
	return &serviceMetrics{
		stream:      stream,
		descriptors: make([]*pb.MetricDescriptor, 0),
	}
}

func (s *serviceMetrics) copyDescriptors() []*pb.MetricDescriptor {
	s.mu.Lock()
	defer s.mu.Unlock()

	descriptors := make([]*pb.MetricDescriptor, len(s.descriptors))
	copy(descriptors, s.descriptors)

	return descriptors
}

func (s *serviceMetrics) addDescriptor(descriptor *pb.MetricDescriptor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.descriptors = append(s.descriptors, descriptor)
}
