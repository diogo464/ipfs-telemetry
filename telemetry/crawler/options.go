package crawler

import (
	"github.com/diogo464/telemetry/walker"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
)

type Option func(*options) error

type options struct {
	logger        *zap.Logger
	meterProvider metric.MeterProvider
	observers     []Observer
	walkerOpts    []walker.Option
}

func WithLogger(l *zap.Logger) Option {
	return func(o *options) error {
		o.logger = l
		return nil
	}
}

func WithMeterProvider(meterProvider metric.MeterProvider) Option {
	return func(o *options) error {
		o.meterProvider = meterProvider
		return nil
	}
}

func WithWalkerObserver(observer walker.Observer) Option {
	return func(o *options) error {
		o.observers = append(o.observers, newWalkerObserverBridge(observer))
		return nil
	}
}

func WithObserver(observer Observer) Option {
	return func(o *options) error {
		o.observers = append(o.observers, observer)
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
		logger:        zap.NewNop(),
		meterProvider: noop.NewMeterProvider(),
		observers:     []Observer{},
		walkerOpts:    []walker.Option{},
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
