package stream

import (
	"time"

	"github.com/diogo464/telemetry/internal/bpool"
)

type streamOptions struct {
	// Maximum size taken up by all segments
	maxSize int
	// Maximum size of a write
	maxWriteSize int
	// Default buffer size allocated, small writes go to the same buffer until it reached defaultBufferSize
	defaultBufferSize int
	// How long until the active buffer turns into a StreamSegment
	activeBufferLifetime time.Duration
	// How long does a segment remain in the stream
	segmentLifetime time.Duration

	// Buffer pool
	bufferPool *bpool.Pool
}

type Option func(*streamOptions)

func streamDefault() *streamOptions {
	return &streamOptions{
		maxSize:              8 * 1024 * 1024,
		maxWriteSize:         1 * 1024 * 1024,
		defaultBufferSize:    4 * 1024,
		activeBufferLifetime: time.Minute * 5,
		segmentLifetime:      time.Minute * 30,
		bufferPool:           nil,
	}
}

func streamApply(o *streamOptions, opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithSegmentLifetime(dur time.Duration) Option {
	return func(so *streamOptions) {
		so.segmentLifetime = dur
	}
}

func WithActiveBufferLifetime(dur time.Duration) Option {
	return func(so *streamOptions) {
		so.activeBufferLifetime = dur
	}
}

func WithPool(pool *bpool.Pool) Option {
	return func(so *streamOptions) {
		so.bufferPool = pool
	}
}
