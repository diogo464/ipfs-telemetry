package telemetry

import (
	"context"
	"fmt"
	"time"
)

var ErrCollectorAlreadyRegistered = fmt.Errorf("collector already registered")

var (
	collectorDefaultPeriod time.Duration = time.Second * 30
)

type CollectorOption func(*collectorConfig) error

type Collector interface {
	// Signal the collector that it should start collecting data.
	Open()
	Descriptor() CollectorDescriptor
	// Drain all data from the collector into the stream.
	Collect(context.Context, *Stream) error
	// Close the collector.
	Close()
}

type CollectorDescriptor struct {
	Name     string
	Period   time.Duration
	Encoding Encoding
}

type collectorConfig struct {
	name     *string
	period   *time.Duration
	encoding *Encoding
}

func collectorConfigDefaults() *collectorConfig {
	return &collectorConfig{
		name:     nil,
		period:   nil,
		encoding: nil,
	}
}

func collectorConfigApply(c *collectorConfig, opts ...CollectorOption) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

func WithCollectorName(name string) CollectorOption {
	return func(c *collectorConfig) error {
		c.name = &name
		return nil
	}
}

func WithCollectorPeriod(period time.Duration) CollectorOption {
	return func(c *collectorConfig) error {
		c.period = &period
		return nil
	}
}

func WithCollectorEncoding(encoding Encoding) CollectorOption {
	return func(c *collectorConfig) error {
		c.encoding = &encoding
		return nil
	}
}

func (s *Service) collectorMainLoop(ctx context.Context, stream *Stream, collector Collector, descriptor CollectorDescriptor) {
	collector.Open()
	defer collector.Close()
	ticker := time.NewTicker(descriptor.Period)

LOOP:
	for {
		select {
		case <-ticker.C:
			latestSeqN := stream.LatestSeqN()
			if err := collector.Collect(ctx, stream); err != nil {
				log.Warnf("collector error[", descriptor.Name, "]: ", err)
			}
			if latestSeqN != stream.LatestSeqN() {
				s.notifyObservers(descriptor.Name)
			}
		case <-ctx.Done():
			break LOOP
		}
	}
}

func (s *Service) notifyObservers(streamName string) {
	entry := s.streams[streamName]
	if entry == nil {
		return
	}

	entry.observers_mu.Lock()
	defer entry.observers_mu.Unlock()
	for observer := range entry.observers {
		select {
		case observer <- struct{}{}:
		default:
		}
	}
}
