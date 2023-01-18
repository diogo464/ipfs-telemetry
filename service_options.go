package telemetry

import (
	"net"
	"time"

	"github.com/multiformats/go-multiaddr"
)

type ServiceOption func(*serviceOptions) error

type serviceOptions struct {
	enableBandwidth      bool
	enableDebug          bool
	defaultStreamOptions []StreamOption
	listener             net.Listener
	metricsPeriod        time.Duration
	enablePush           bool
	pushTargets          []multiaddr.Multiaddr
	pushInterval         time.Duration
}

func serviceDefaults() *serviceOptions {
	return &serviceOptions{
		enableBandwidth:      false,
		enableDebug:          false,
		defaultStreamOptions: []StreamOption{},
		listener:             nil,
		metricsPeriod:        time.Second * 15,
		enablePush:           false,
		pushTargets:          []multiaddr.Multiaddr{},
		pushInterval:         time.Minute * 15,
	}
}

func serviceApply(o *serviceOptions, os ...ServiceOption) error {
	for _, opt := range os {
		err := opt(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func WithServiceBandwidth(enabled bool) ServiceOption {
	return func(so *serviceOptions) error {
		so.enableBandwidth = enabled
		return nil
	}
}

func WithServiceDebug(enabled bool) ServiceOption {
	return func(so *serviceOptions) error {
		so.enableDebug = enabled
		return nil
	}
}

func WithServiceDefaultStreamOpts(opts ...StreamOption) ServiceOption {
	return func(so *serviceOptions) error {
		so.defaultStreamOptions = opts
		return nil
	}
}

func WithServiceListener(listener net.Listener) ServiceOption {
	return func(so *serviceOptions) error {
		so.listener = listener
		return nil
	}
}

func WithServiceTcpListener(addr string) ServiceOption {
	return func(so *serviceOptions) error {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		so.listener = listener
		return nil
	}
}

func WithServiceMetricsPeriod(period time.Duration) ServiceOption {
	return func(so *serviceOptions) error {
		so.metricsPeriod = period
		return nil
	}
}

func WithServicePush(enabled bool) ServiceOption {
	return func(so *serviceOptions) error {
		so.enablePush = enabled
		return nil
	}
}

func WithServicePushTargets(targets ...multiaddr.Multiaddr) ServiceOption {
	return func(so *serviceOptions) error {
		so.pushTargets = targets
		return nil
	}
}
