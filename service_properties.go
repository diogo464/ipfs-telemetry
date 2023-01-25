package telemetry

import (
	"sync"

	"github.com/diogo464/telemetry/internal/pb"
)

type propertyId struct {
	scope string
	name  string
}

type serviceProperties struct {
	mu          sync.Mutex
	propertyIds map[propertyId]struct{}
	properties  []*pb.Property
	descriptors []*pb.PropertyDescriptor
	nextId      uint32
}

func newServiceProperties() *serviceProperties {
	return &serviceProperties{
		propertyIds: make(map[propertyId]struct{}),
		properties:  make([]*pb.Property, 0),
		descriptors: make([]*pb.PropertyDescriptor, 0),
	}
}

func (s *serviceProperties) copyDescriptors() []*pb.PropertyDescriptor {
	s.mu.Lock()
	defer s.mu.Unlock()

	descriptors := make([]*pb.PropertyDescriptor, len(s.descriptors))
	copy(descriptors, s.descriptors)

	return descriptors
}

func (s *serviceProperties) copyProperties() []*pb.Property {
	s.mu.Lock()
	defer s.mu.Unlock()

	properties := make([]*pb.Property, len(s.properties))
	copy(properties, s.properties)

	return properties
}

func (s *serviceProperties) create(config propertyConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid := propertyId{
		scope: config.Scope,
		name:  config.Name,
	}
	if _, ok := s.propertyIds[pid]; ok {
		return
	}

	id := s.nextId
	s.propertyIds[pid] = struct{}{}
	s.nextId++

	// TODO: maybe remove property descriptors
	s.descriptors = append(s.descriptors, &pb.PropertyDescriptor{
		Id:          id,
		Scope:       config.Scope,
		Name:        config.Name,
		Description: config.Description,
	})
	s.properties = append(s.properties, propertyConfigToPb(id, config))
}
