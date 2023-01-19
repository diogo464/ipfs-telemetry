package telemetry

import (
	"github.com/diogo464/telemetry/internal/pb"
)

func (s *Service) addMetricDescriptor(desc MetricDescriptor) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.descriptors = append(s.metrics.descriptors, &pb.MetricDescriptor{
		Scope:       desc.Scope,
		Name:        desc.Name,
		Description: desc.Description,
	})
}

func (s *Service) createCapture(c captureConfig) {
	captures := s.captures
	captures.mu.Lock()
	defer captures.mu.Unlock()

	id := captures.nextID
	captures.nextID++

	if _, ok := captures.captures[id]; ok {
		log.Warnf("capture already exists", "capture", c.Name)
		return
	}

	captures.captures[id] = newServiceCapture(s.Context(), id, s.newStream(), c)
}

func (s *Service) createEvent(config eventConfig) EventEmitter {
	events := s.events
	s.events.mu.Lock()
	defer s.events.mu.Unlock()

	id := events.nextID
	events.nextID++

	if e, ok := events.events[id]; ok {
		return e.emitter
	}

	em := newEventEmitter(config, s.newStream())
	events.events[id] = &serviceEvent{
		config:  config,
		stream:  em.stream,
		emitter: em,
		descriptor: &pb.EventDescriptor{
			Id:          id,
			Scope:       config.Scope,
			Name:        config.Name,
			Description: config.Description,
		},
	}

	return em
}

func (s *Service) createProperty(c propertyConfig) {
	s.properties.mu.Lock()
	defer s.properties.mu.Unlock()

	for _, prop := range s.properties.properties {
		if prop.pbproperty.Name == c.Name {
			log.Warnf("property already exists", "property", c.Name)
			return
		}
	}

	id := uint32(len(s.properties.properties))
	pbprop := &pb.Property{
		Id:          id,
		Scope:       c.Scope,
		Name:        c.Name,
		Description: c.Description,
	}

	s.properties.properties = append(s.properties.properties, &servicePropertyEntry{pbproperty: pbprop})
}
