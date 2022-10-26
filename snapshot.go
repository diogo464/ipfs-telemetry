package telemetry

import (
	"encoding/json"
	"fmt"
	"time"
)

var ErrSnapshotAlreadyRegistered = fmt.Errorf("snapshot already registered")

var _ (SnapshotCollector) = (*Snapshot)(nil)

type SnapshotCollector interface {
	Descriptor() SnapshotDescriptor
	Collect(*Stream) error
}

type SnapshotDescriptor struct {
	Name   string
	Period time.Duration
}

type SnapshotConfig struct {
	Name      string
	Period    time.Duration
	Collector func() (interface{}, error)
}

type Snapshot struct {
	descriptor SnapshotDescriptor
	collector  func() (interface{}, error)
}

func NewSnapshot(config SnapshotConfig) *Snapshot {
	if config.Collector == nil {
		panic("nil collector given to NewSnapshot")
	}

	return &Snapshot{
		descriptor: SnapshotDescriptor{
			Name:   config.Name,
			Period: config.Period,
		},
		collector: config.Collector,
	}
}

// Descriptor implements SnapshotCollector
func (s *Snapshot) Descriptor() SnapshotDescriptor {
	return s.descriptor
}

// Collect implements SnapshotCollector
func (s *Snapshot) Collect(stream *Stream) error {
	data, err := s.collector()
	if err != nil {
		return err
	}
	marshaled, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return stream.Write(marshaled)
}
