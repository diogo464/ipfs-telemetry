package monitor

import (
	"context"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

const (
	DEFAULT_MAX_FAILED_ATTEMPTS = 3
	DEFAULT_RETRY_INTERVAL      = time.Second * 30
	DEFAULT_COLLECT_PERIOD      = time.Minute * 2
	DEFAULT_BANDWIDTH_PERIOD    = time.Minute * 30
)

type Option func(*options) error

type options struct {
	// How many consecutive errors can happen while making requests
	// to a peer before that peer is removed
	MaxFailedAttemps int
	// How long before retrying a request to a peer after a failure
	RetryInterval time.Duration
	// How often should telemetry be collected from peers
	CollectPeriod   time.Duration
	BandwidthPeriod time.Duration
	Host            host.Host
}

func defaults() *options {
	return &options{
		MaxFailedAttemps: DEFAULT_MAX_FAILED_ATTEMPTS,
		RetryInterval:    DEFAULT_RETRY_INTERVAL,
		CollectPeriod:    DEFAULT_COLLECT_PERIOD,
		BandwidthPeriod:  DEFAULT_BANDWIDTH_PERIOD,
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

func WithCollectPeriod(period time.Duration) Option {
	return func(o *options) error {
		o.CollectPeriod = period
		return nil
	}
}

func WithBandwidthPeriod(period time.Duration) Option {
	return func(o *options) error {
		o.BandwidthPeriod = period
		return nil
	}
}

func WithHost(h host.Host) Option {
	return func(o *options) error {
		o.Host = h
		return nil
	}
}

func createDefaultHost(ctx context.Context) (host.Host, error) {
	return libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			client := dht.NewDHTClient(ctx, h, datastore.NewMapDatastore())
			if err := client.Bootstrap(ctx); err != nil {
				return nil, err
			}

			var err error = nil
			var success bool = false
			for _, bootstrap := range dht.GetDefaultBootstrapPeerAddrInfos() {
				err = h.Connect(ctx, bootstrap)
				if err == nil {
					success = true
				}
			}

			if success {
				client.RefreshRoutingTable()
				time.Sleep(time.Second * 2)
				return client, nil
			} else {
				return nil, err
			}
		}))
}
