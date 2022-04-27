package probe

import "time"

const (
	DEFAULT_PROBE_NAME                    string        = "probe"
	DEFAULT_PROBE_NEW_SESSION_INTERVAL    time.Duration = time.Second * 10
	DEFAULT_PROBE_SESSION_PROVIDERS_LIMIT int           = 64
	DEFAULT_PROBE_SESSION_LIFETIME_LIMIT  time.Duration = time.Minute * 5
	DEFAULT_PROBE_MAX_ONGOING             int           = 128
)

type Option func(*options) error

type options struct {
	probeName                  string
	probeNewSessionInterval    time.Duration
	probeSessionProvidersLimit int
	probeSessionLifetimeLimit  time.Duration
	probeMaxOngoing            int
}

func defaults() *options {
	return &options{
		probeName:                  DEFAULT_PROBE_NAME,
		probeNewSessionInterval:    DEFAULT_PROBE_NEW_SESSION_INTERVAL,
		probeSessionProvidersLimit: DEFAULT_PROBE_SESSION_PROVIDERS_LIMIT,
		probeSessionLifetimeLimit:  DEFAULT_PROBE_SESSION_LIFETIME_LIMIT,
		probeMaxOngoing:            DEFAULT_PROBE_MAX_ONGOING,
	}
}

func apply(opts *options, o ...Option) error {
	for _, f := range o {
		if err := f(opts); err != nil {
			return err
		}
	}
	return nil
}

func WithName(name string) Option {
	return func(o *options) error {
		o.probeName = name
		return nil
	}
}

func WithNewSessionInterval(interval time.Duration) Option {
	return func(o *options) error {
		o.probeNewSessionInterval = interval
		return nil
	}
}

func WithSessionProvidersLimit(limit int) Option {
	return func(o *options) error {
		o.probeSessionProvidersLimit = limit
		return nil
	}
}

func WithSessionLifetimeLimit(limit time.Duration) Option {
	return func(o *options) error {
		o.probeSessionLifetimeLimit = limit
		return nil
	}
}

func WithMaxOngoing(max int) Option {
	return func(o *options) error {
		o.probeMaxOngoing = max
		return nil
	}
}
