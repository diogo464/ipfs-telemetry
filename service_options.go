package telemetry

import (
	"net"
	"time"

	"github.com/multiformats/go-multiaddr"
	"go.opentelemetry.io/otel/sdk/resource"
)

type ServiceOption func(*serviceOptions) error

type serviceOptions struct {
	enableBandwidth      bool
	enableDebug          bool
	listener             net.Listener
	metricsPeriod        time.Duration
	windowDuration       time.Duration
	activeBufferDuration time.Duration
	enablePush           bool
	pushTargets          []multiaddr.Multiaddr
	pushInterval         time.Duration
	otelResource         *resource.Resource
}

func serviceDefaults() *serviceOptions {
	return &serviceOptions{
		enableBandwidth:      false,
		enableDebug:          false,
		listener:             nil,
		metricsPeriod:        time.Second * 15,
		windowDuration:       time.Minute * 30,
		activeBufferDuration: time.Minute * 5,
		enablePush:           false,
		pushTargets:          []multiaddr.Multiaddr{},
		pushInterval:         time.Minute * 15,
		otelResource:         nil,
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

func WithServiceWindowDuration(duration time.Duration) ServiceOption {
	return func(so *serviceOptions) error {
		so.windowDuration = duration
		return nil
	}
}

func WithServiceActiveBufferDuration(duration time.Duration) ServiceOption {
	return func(so *serviceOptions) error {
		so.activeBufferDuration = duration
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

func WithServiceResource(resource *resource.Resource) ServiceOption {
	return func(so *serviceOptions) error {
		so.otelResource = resource
		return nil
	}
}
