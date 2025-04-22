package bpool

import (
	"github.com/diogo464/telemetry/internal/utils"
)

type options struct {
	allocSize int
	maxSize   int
}

type Option func(*options)

func defaults() *options {
	return &options{
		allocSize: 1 << 16,
		maxSize:   1 << 30,
	}
}

// Size of each buffer allocated by the pool.
// This value will be rounded up to the nearest power of two.
// Any buffer request that is larger than this value will be allocated
// separately and not pooled.
func WithAllocSize(size int) Option {
	return func(o *options) {
		o.allocSize = utils.RoundUpPowerOfTwo(size)
	}
}

// Maximum size held by the pool.
// If a buffer request is made that would exceed this size,
// the buffer will be allocated separately and not pooled.
// This value will be rounded up to the nearest power of two.
func WithMaxSize(size int) Option {
	return func(o *options) {
		o.maxSize = utils.RoundUpPowerOfTwo(size)
	}
}
