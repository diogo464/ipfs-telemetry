package collector

import (
	"context"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/measurements"
	"go.uber.org/atomic"
)

var _ Collector = (*HolePunchCollector)(nil)
var _ measurements.HolePunch = (*HolePunchCollector)(nil)

type HolePunchCollector struct {
	success *atomic.Uint32
	failure *atomic.Uint32
}

func NewHolePunchCollector() Collector {
	c := &HolePunchCollector{
		success: atomic.NewUint32(0),
		failure: atomic.NewUint32(0),
	}
	measurements.HolePunchRegister(c)
	return c
}

// Close implements Collector
func (*HolePunchCollector) Close() {
}

// Collect implements Collector
func (c *HolePunchCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	sink.Push(&datapoint.HolePunch{
		Timestamp: datapoint.NewTimestamp(),
		Success:   c.success.Load(),
		Failure:   c.failure.Load(),
	})
}

// Register implements measurements.HolePunch
func (c *HolePunchCollector) Register(success bool) {
	if success {
		c.success.Inc()
	} else {
		c.failure.Inc()
	}
}
