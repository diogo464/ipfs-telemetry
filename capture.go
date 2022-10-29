package telemetry

import (
	"context"
	"encoding/json"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
)

type serviceCapture struct {
	pbdescriptor *pb.CaptureDescriptor
	stream       *Stream
	callback     CaptureCallback
}

func newServiceCapture(ctx context.Context, stream *Stream, config CaptureConfig) *serviceCapture {
	sc := &serviceCapture{
		pbdescriptor: &pb.CaptureDescriptor{
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
