package monitor

import (
	"context"
	"net"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

const (
	DEFAULT_MAX_FAILED_ATTEMPTS = 15
	DEFAULT_RETRY_INTERVAL      = time.Second * 30
	DEFAULT_COLLECT_ENABLED     = true
	DEFAULT_COLLECT_PERIOD      = time.Minute * 5
	DEFAULT_COLLECT_TIMEOUT     = time.Minute * 2
	DEFAULT_BANDWIDTH_ENABLED   = true
	DEFAULT_BANDWIDTH_PERIOD    = time.Minute * 30
	DEFAULT_BANDWIDTH_TIMEOUT   = time.Minute * 5
)

type Option func(*options) error

type options struct {
	// How many consecutive errors can happen while making requests
	// to a peer before that peer is removed
	MaxFailedAttemps int
	// How long before retrying a request to a peer after a failure
	RetryInterval time.Duration
	// How often should telemetry be collected from peers
	CollectEnabled   bool
	CollectPeriod    time.Duration
	CollectTimeout   time.Duration
	BandwidthEnabled bool
	BandwidthPeriod  time.Duration
	BandwidthTimeout time.Duration
	Host             host.Host
	Exporter         Exporter
	Listener         net.Listener
	Logger           *zap.Logger
	MeterProvider    metric.MeterProvider
}

func defaults() *options {
	return &options{
		MaxFailedAttemps: DEFAULT_MAX_FAILED_ATTEMPTS,
		RetryInterval:    DEFAULT_RETRY_INTERVAL,
		CollectEnabled:   DEFAULT_COLLECT_ENABLED,
		CollectPeriod:    DEFAULT_COLLECT_PERIOD,
		CollectTimeout:   DEFAULT_COLLECT_TIMEOUT,
		BandwidthEnabled: DEFAULT_BANDWIDTH_ENABLED,
		BandwidthPeriod:  DEFAULT_BANDWIDTH_PERIOD,
		BandwidthTimeout: DEFAULT_BANDWIDTH_TIMEOUT,
		Listener:         nil,
		Logger:           zap.NewNop(),
		MeterProvider:    metric.NewNoopMeterProvider(),
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

func WithMaxFailedAttempts(attemps int) Option {
	return func(o *options) error {
		o.MaxFailedAttemps = attemps
		return nil
	}
}

func WithRetryInterval(interval time.Duration) Option {
	return func(o *options) error {
		o.RetryInterval = interval
		return nil
	}
}

func WithCollectEnabled(enabled bool) Option {
	return func(o *options) error {
		o.CollectEnabled = enabled
		return nil
	}
}

func WithCollectPeriod(period time.Duration) Option {
	return func(o *options) error {
		o.CollectPeriod = period
		return nil
	}
}

func WithCollectTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		o.CollectTimeout = timeout
		return nil
	}
}

func WithBandwidthEnabled(enabled bool) Option {
	return func(o *options) error {
		o.BandwidthEnabled = enabled
		return nil
	}
}

func WithBandwidthPeriod(period time.Duration) Option {
	return func(o *options) error {
		o.BandwidthPeriod = period
		return nil
	}
}

func WithBandwidthTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		o.BandwidthTimeout = timeout
		return nil
	}
}

func WithHost(h host.Host) Option {
	return func(o *options) error {
		o.Host = h
		return nil
	}
}

func WithExporter(e Exporter) Option {
	return func(o *options) error {
		o.Exporter = e
		return nil
	}
}

func WithListener(l net.Listener) Option {
	return func(o *options) error {
		o.Listener = l
		return nil
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(o *options) error {
		o.Logger = l
		return nil
	}
}

func WithMeterProvider(m metric.MeterProvider) Option {
	return func(o *options) error {
		o.MeterProvider = m
		return nil
	}
}

func createDefaultHost(ctx context.Context) (host.Host, error) {
	return libp2p.New(
		libp2p.NoListenAddrs,
		// libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		// 	client := dht.NewDHTClient(ctx, h, datastore.NewMapDatastore())
		// 	if err := client.Bootstrap(ctx); err != nil {
		// 		return nil, err
		// 	}

		// 	var err error = nil
		// 	var success bool = false
		// 	for _, bootstrap := range dht.GetDefaultBootstrapPeerAddrInfos() {
		// 		err = h.Connect(ctx, bootstrap)
		// 		if err == nil {
		// 			success = true
		// 		}
		// 	}

		// 	if success {
		// 		client.RefreshRoutingTable()
		// 		time.Sleep(time.Second * 2)
		// 		return client, nil
		// 	} else {
		// 		return nil, err
		// 	}
		// })
	)
}
