package telemetry

import "time"

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
}

type StreamOption func(*streamOptions)

func streamDefault() *streamOptions {
	return &streamOptions{
		maxSize:              8 * 1024 * 1024,
		maxWriteSize:         1 * 1024 * 1024,
		defaultBufferSize:    128 * 1024,
		activeBufferLifetime: time.Minute * 5,
		segmentLifetime:      time.Minute * 30,
	}
}

func streamApply(o *streamOptions, opts ...StreamOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithStreamSegmentLifetime(dur time.Duration) StreamOption {
	return func(so *streamOptions) {
		so.segmentLifetime = dur
	}
}

func WithStreamActiveBufferLifetime(dur time.Duration) StreamOption {
	return func(so *streamOptions) {
		so.activeBufferLifetime = dur
	}
}
