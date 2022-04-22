package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/datapoint"
)

type Collector interface {
	Collect(context.Context, datapoint.Sink)
	Close()
}

func RunCollector(ctx context.Context, interval time.Duration, sink datapoint.Sink, collector Collector) {
	go func() {
		ticker := time.NewTicker(interval)

	LOOP:
		for {
			select {
			case <-ticker.C:
				collector.Collect(ctx, sink)
			case <-ctx.Done():
				break LOOP
			}
		}
	}()
}
