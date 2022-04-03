package collector

import (
	"context"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
)

type KademliaOptions struct {
	Interval time.Duration
}

func RunKademliaCollector(ctx context.Context, sink snapshot.Sink, opts KademliaOptions) {
	ticker := time.NewTicker(opts.Interval)

LOOP:
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			break LOOP
		}
	}
}
