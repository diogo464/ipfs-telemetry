package walker

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type Option func(*options) error

type options struct {
	connectTimeout time.Duration
	requestTimeout time.Duration
	interval       time.Duration
	concurrency    uint
	seeds          []peer.AddrInfo
	observer       Observer
}

func WithConnectTimeout(timeout time.Duration) Option {
	return func(o *options) error {
		o.connectTimeout = timeout
		return nil
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(c *options) error {
		c.requestTimeout = timeout
		return nil
	}
}

func WithInterval(interval time.Duration) Option {
	return func(c *options) error {
		c.interval = interval
		return nil
	}
}

// How many requests can be happening in parallel
func WithConcurrency(concurrency uint) Option {
	return func(c *options) error {
		if concurrency == 0 {
			concurrency = 1
		}
		c.concurrency = concurrency
		return nil
	}
}

func WithSeeds(seeds []peer.AddrInfo) Option {
	return func(c *options) error {
		c.seeds = seeds
		return nil
	}
}

func WithObserver(observer Observer) Option {
	return func(c *options) error {
		c.observer = observer
		return nil
	}
}

func defaults(c *options) {
	c.connectTimeout = time.Second * 5
	c.requestTimeout = time.Second * 25
	c.interval = time.Millisecond * 20
	c.concurrency = 128
	c.seeds = dht.GetDefaultBootstrapPeerAddrInfos()
	c.observer = &NullObserver{}
}

func apply(c *options, opts ...Option) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}
