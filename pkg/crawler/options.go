package crawler

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type Option func(*config) error

type config struct {
	requestTimeout time.Duration
	concurrency    uint
	seeds          []peer.AddrInfo
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *config) error {
		c.requestTimeout = timeout
		return nil
	}
}

// How many requests can be happening in parallel
func WithConcurrency(concurrency uint) Option {
	return func(c *config) error {
		if concurrency == 0 {
			concurrency = 1
		}
		c.concurrency = concurrency
		return nil
	}
}

func WithSeeds(seeds []peer.AddrInfo) Option {
	return func(c *config) error {
		c.seeds = seeds
		return nil
	}
}

func defaults(c *config) {
	c.requestTimeout = time.Minute * 2
	c.concurrency = 96
	c.seeds = dht.GetDefaultBootstrapPeerAddrInfos()
}

func apply(c *config, opts ...Option) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}
