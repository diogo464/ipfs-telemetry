package telemetry

import "time"

type Option func(*options) error

type options struct {
	// a string defining the how long datapoint. are kept.
	// format is the same one used by time.ParseDuration
	windowDuration time.Duration
}

var defaults = func(o *options) {
	o.windowDuration = time.Minute * 90
}

func WithWindowDuration(duration time.Duration) Option {
	return func(o *options) error {
		o.windowDuration = duration
		return nil
	}
}
