package collector

import (
	"context"
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/window"
)

var _ Collector = (*windowCollector)(nil)

type windowCollector struct {
	wnd      window.Window
	stats    *window.Stats
	duration time.Duration
}

func NewWindowCollector(d time.Duration, wnd window.Window) Collector {
	return &windowCollector{
		wnd:      wnd,
		stats:    new(window.Stats),
		duration: d,
	}
}

// Close implements Collector
func (*windowCollector) Close() {
}

// Collect implements Collector
func (c *windowCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	c.wnd.Stats(c.stats)
	sink.Push(&datapoint.Window{
		Timestamp:       datapoint.NewTimestamp(),
		WindowDuration:  c.duration,
		DatapointCount:  c.stats.Count,
		DatapointMemory: c.stats.Memory,
	})
}
