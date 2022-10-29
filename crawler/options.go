package crawler

import "github.com/diogo464/telemetry/walker"

type Option func(*options) error

type options struct {
	observer    walker.Observer
	concurrency int
}

func WithObserver(observer walker.Observer) Option {
	return func(o *options) error {
		o.observer = observer
		return nil
	}
}

func WithConcurrency(concurrency int) Option {
	return func(o *options) error {
		o.concurrency = concurrency
		return nil
	}
}

func defaults() *options {
	return &options{
		observer: &walker.NullObserver{},
	}
}

func apply(opts *options, o ...Option) error {
	for _, opt := range o {
		if err := opt(opts); err != nil {
			return err
		}
	}
	return nil
}
