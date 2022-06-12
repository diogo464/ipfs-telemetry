package telemetry

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

var ErrCollectorAlreadyRegistered = fmt.Errorf("collector already registered")

type Collector interface {
	Name() string
	Collect(context.Context, *Stream) error
	Close()
}

type CollectorOpts struct {
	Interval time.Duration
}

func collectorMainLoop(ctx context.Context, stream *Stream, collector Collector, opts CollectorOpts) {
	defer collector.Close()
	ticker := time.NewTicker(opts.Interval)

LOOP:
	for {
		select {
		case <-ticker.C:
			if err := collector.Collect(ctx, stream); err != nil {
				logrus.Error("collector error[", collector.Name(), "]: ", err)
			}
		case <-ctx.Done():
			break LOOP
		}
	}
}
