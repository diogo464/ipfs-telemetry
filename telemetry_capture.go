package telemetry

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
)

type CaptureCallback func(context.Context) (interface{}, error)

type CaptureDescriptor struct {
	StreamId    StreamId `json:"stream_id"`
	Scope       string   `json:"scope"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

type Capture struct {
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
}

type captureConfig struct {
	Scope       string
	Name        string
	Description string
	Callback    CaptureCallback
	Interval    time.Duration
}

func captureDescriptorToPb(descriptor CaptureDescriptor) *pb.CaptureDescriptor {
	return &pb.CaptureDescriptor{
		StreamId:    uint32(descriptor.StreamId),
		Scope:       descriptor.Scope,
		Name:        descriptor.Name,
		Description: descriptor.Description,
	}
}

func captureDescriptorFromPb(descriptor *pb.CaptureDescriptor) CaptureDescriptor {
	return CaptureDescriptor{
		StreamId:    StreamId(descriptor.StreamId),
		Scope:       descriptor.Scope,
		Name:        descriptor.Name,
		Description: descriptor.Description,
	}
}
