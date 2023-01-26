package crawler

import "github.com/diogo464/telemetry/walker"

type Option func(*options) error

type options struct {
	observer   walker.Observer
	walkerOpts []walker.Option
}

func WithObserver(observer walker.Observer) Option {
	return func(o *options) error {
		o.observer = observer
		return nil
	}
}

func WithWalkerOption(walkerOpt ...walker.Option) Option {
	return func(o *options) error {
		o.walkerOpts = append(o.walkerOpts, walkerOpt...)
		return nil
	}
}

func defaults() *options {
	return &options{
		observer:   &walker.NullObserver{},
		walkerOpts: []walker.Option{},
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
