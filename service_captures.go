package telemetry

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
)

type captureId struct {
	scope       string
	name        string
	description string
}

type serviceCapture struct {
	stream   *stream.Stream
	callback CaptureCallback
}

type serviceCaptures struct {
	ctx     context.Context
	streams *serviceStreams

	mu          sync.Mutex
	descriptors []*pb.CaptureDescriptor
	captures    map[captureId]*serviceStream
}

func newServiceCaptures(ctx context.Context, streams *serviceStreams) *serviceCaptures {
	return &serviceCaptures{
		ctx:     ctx,
		streams: streams,

		descriptors: make([]*pb.CaptureDescriptor, 0),
		captures:    make(map[captureId]*serviceStream),
	}
}

func (c *serviceCaptures) copyDescriptors() []*pb.CaptureDescriptor {
	c.mu.Lock()
	defer c.mu.Unlock()

	descriptors := make([]*pb.CaptureDescriptor, len(c.descriptors))
	copy(descriptors, c.descriptors)

	return descriptors
}

func (c *serviceCaptures) create(config captureConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := captureId{
		scope:       config.Scope,
		name:        config.Name,
		description: config.Description,
	}
	if ss, ok := c.captures[id]; ok {
		newServiceCapture(c.ctx, ss.stream, config)
		return
	}

	stream := c.streams.create()
	descriptor := CaptureDescriptor{
		StreamId:    stream.streamId,
		Scope:       config.Scope,
		Name:        config.Name,
		Description: config.Description,
	}
	c.descriptors = append(c.descriptors, captureDescriptorToPb(descriptor))
	c.captures[id] = stream
}

func newServiceCapture(ctx context.Context, stream *stream.Stream, config captureConfig) *serviceCapture {
	sc := &serviceCapture{
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

		data, err := c.callback(ctx)
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
