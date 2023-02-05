package telemetry

import (
	"sync"

	"github.com/diogo464/telemetry/internal/pb"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

type propertyId struct {
	scope       instrumentation.Scope
	name        string
	description string
}

type serviceProperties struct {
	mu          sync.Mutex
	propertyIds map[propertyId]struct{}
	properties  []*pb.Property
}

func newServiceProperties() *serviceProperties {
	return &serviceProperties{
		propertyIds: make(map[propertyId]struct{}),
		properties:  make([]*pb.Property, 0),
	}
}

func (s *serviceProperties) copyProperties() []*pb.Property {
	s.mu.Lock()
	defer s.mu.Unlock()

	properties := make([]*pb.Property, len(s.properties))
	copy(properties, s.properties)

	return properties
}

func (s *serviceProperties) create(prop Property) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid := propertyId{
		scope:       prop.Scope,
		name:        prop.Name,
		description: prop.Description,
	}
	if _, ok := s.propertyIds[pid]; ok {
		return
	}

	s.propertyIds[pid] = struct{}{}
	proppb := &pb.Property{
		Scope: &v1.InstrumentationScope{
			Name:    prop.Scope.Name,
			Version: prop.Scope.Version,
		},
		Name:        prop.Name,
		Description: prop.Description,
	}

	switch v := prop.Value.(type) {
	case *PropertyValueInteger:
		proppb.Value = &pb.Property_IntegerValue{
			IntegerValue: v.GetInteger(),
		}
	case *PropertyValueString:
		proppb.Value = &pb.Property_StringValue{
			StringValue: v.GetString(),
		}
	default:
	}

	s.properties = append(s.properties, proppb)
}

func (s *serviceProperties) getSize() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.properties)
}
