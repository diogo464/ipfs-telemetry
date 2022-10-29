package telemetry

import "github.com/diogo464/telemetry/internal/pb"

// Capture implements Telemetry
func (s *Service) Capture(c CaptureConfig) {
	captures := s.captures
	captures.mu.Lock()
	defer captures.mu.Unlock()

	if _, ok := captures.captures[c.Name]; ok {
		log.Warnf("capture already exists", "capture", c.Name)
		return
	}

	captures.captures[c.Name] = newServiceCapture(s.Context(), NewStream(s.opts.defaultStreamOptions...), c)
}

// Event implements Telemetry
func (s *Service) Event(config EventConfig) EventEmitter {
	events := s.events
	s.events.mu.Lock()
	defer s.events.mu.Unlock()

	if e, ok := events.events[config.Name]; ok {
		return e.emitter
	}

	em := newEventEmitter(config, NewStream(s.opts.defaultStreamOptions...))
	events.events[config.Name] = &serviceEvent{
		config:  config,
		stream:  em.stream,
		emitter: em,
		descriptor: &pb.EventDescriptor{
			Name:        config.Name,
			Description: config.Description,
		},
	}

	return em
}

// PropertyInt implements Telemetry
func (s *Service) Property(c PropertyConfig) {
	s.properties.mu.Lock()
	defer s.properties.mu.Unlock()

	pbprop := propertyConfigToPb(c)
	s.addProperty(pbprop)
}

func (s *Service) addProperty(p *pb.Property) {
	properties := s.properties
	if _, ok := properties.properties[p.Name]; ok {
		log.Warnf("property already exists", "property", p.Name)
		return
	}

	properties.properties[p.Name] = &servicePropertyEntry{
		pbproperty: p,
	}
}
