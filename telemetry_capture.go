package telemetry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
)

type CaptureCallback func(context.Context) (interface{}, error)

type CaptureDescriptor struct {
	ID          uint32
	Scope       string
	Name        string
	Description string
}

type Capture struct {
	Timestamp time.Time
	Data      []byte
}

type captureConfig struct {
	Scope       string
	Name        string
	Description string
	Callback    CaptureCallback
	Interval    time.Duration
}

type serviceCapture struct {
	pbdescriptor *pb.CaptureDescriptor
	stream       *stream.Stream
	callback     CaptureCallback
}

func newServiceCapture(ctx context.Context, id uint32, stream *stream.Stream, config captureConfig) *serviceCapture {
	sc := &serviceCapture{
		pbdescriptor: &pb.CaptureDescriptor{
			Id:          id,
			Scope:       config.Scope,
			Name:        config.Name,
			Description: config.Description,
		},
		stream:   stream,
		callback: config.Callback,
	}
	go sc.captureTask(ctx, config.Name, config.Interval)
	return sc
}

func (c *serviceCapture) captureTask(ctx context.Context, name string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		data, err := c.callback(context.TODO())
		if err != nil {
			log.Warn("capture callback failed", "capture", name, "error", err)
			continue
		}

		marshaled, err := json.Marshal(data)
		if err != nil {
			log.Warn("failed to marshal capture data", "capture", name, "error", err)
			continue
		}

		if err := c.stream.Write(marshaled); err != nil {
			log.Error("failed to write capture to stream", "capture", name, "error", err)
		}
	}
}
