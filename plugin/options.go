package telemetry

import "time"

type Option func(*options) error

type options struct {
	// a string defining the how long snapshots are kept.
	// format is the same one used by time.ParseDuration
	windowDuration time.Duration
}

var defaults = func(o *options) error {
	o.windowDuration = time.Minute * 30

	return nil
}

func WithWindowDuration(duration time.Duration) Option {
	return func(o *options) error {
		o.windowDuration = duration
		return nil
	}
}
