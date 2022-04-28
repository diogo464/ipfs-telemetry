package orchestrator

import "net"

const DEFAULT_NUM_CIDS int = 16

type Option = func(*options) error

type options struct {
	exporter   Exporter
	probeAddrs []net.Addr
	numCids    int
}

func defaults() *options {
	return &options{
		exporter:   &NullExporter{},
		probeAddrs: []net.Addr{},
		numCids:    DEFAULT_NUM_CIDS,
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

func WithExporter(exporter Exporter) Option {
	return func(o *options) error {
		o.exporter = exporter
		return nil
	}
}

func WithProbes(probes []net.Addr) Option {
	return func(o *options) error {
		o.probeAddrs = probes
		return nil
	}
}

func WithNumCids(n int) Option {
	return func(o *options) error {
		o.numCids = n
		return nil
	}
}
