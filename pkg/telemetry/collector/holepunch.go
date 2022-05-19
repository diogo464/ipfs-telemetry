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
	incoming_success *atomic.Uint32
	incoming_failure *atomic.Uint32
	outgoing_success *atomic.Uint32
	outgoing_failure *atomic.Uint32
}

func NewHolePunchCollector() Collector {
	c := &HolePunchCollector{
		incoming_success: atomic.NewUint32(0),
		incoming_failure: atomic.NewUint32(0),
		outgoing_success: atomic.NewUint32(0),
		outgoing_failure: atomic.NewUint32(0),
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
		Timestamp:       datapoint.NewTimestamp(),
		IncomingSuccess: c.incoming_success.Load(),
		IncomingFailure: c.incoming_failure.Load(),
		OutgoingSuccess: c.outgoing_success.Load(),
		OutgoingFailure: c.outgoing_failure.Load(),
	})
}

// Incoming implements measurements.HolePunch
func (c *HolePunchCollector) Incoming(success bool) {
	if success {
		c.incoming_success.Inc()
	} else {
		c.incoming_failure.Inc()
	}
}

// Outgoing implements measurements.HolePunch
func (c *HolePunchCollector) Outgoing(success bool) {
	if success {
		c.outgoing_success.Inc()
	} else {
		c.outgoing_failure.Inc()
	}
}
