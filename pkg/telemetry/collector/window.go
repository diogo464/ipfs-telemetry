package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry/window"
)

type WindowOptions struct {
	Interval time.Duration
}

func RunWindowCollector(ctx context.Context, d time.Duration, w window.Window, sink snapshot.Sink, opts StorageOptions) {
	ticker := time.NewTicker(opts.Interval)
	stats := new(window.Stats)

	for {
		select {
		case <-ticker.C:
			w.Stats(stats)
			sink.Push(&snapshot.Window{
				Timestamp:      snapshot.NewTimestamp(),
				WindowDuration: d,
				SnapshotCount:  stats.Count,
				SnapshotMemory: stats.Memory,
			})
		case <-ctx.Done():
		}
	}
}
