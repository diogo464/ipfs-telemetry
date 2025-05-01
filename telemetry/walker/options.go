package walker

import (
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
)

type Option func(*options) error

type options struct {
	host           host.Host
	connectTimeout time.Duration
	requestTimeout time.Duration
	interval       time.Duration
	concurrency    uint
	seeds          []peer.AddrInfo
	observer       Observer
	addrFilter     AddressFilter
}

func WithHost(h host.Host) Option {
	return func(o *options) error {
		o.host = h
		return nil
	}
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

func WithAddressFilter(filter AddressFilter) Option {
	return func(c *options) error {
		c.addrFilter = filter
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
	c.addrFilter = AddressFilterPublic
}

func apply(c *options, opts ...Option) error {
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

func newDefaultHost() (host.Host, error) {
	limits := rcmgr.InfiniteLimits
	limiter := rcmgr.NewFixedLimiter(limits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, err
	}
	h, err := libp2p.New(libp2p.NoListenAddrs, libp2p.EnableRelay(), libp2p.ResourceManager(rm))
	if err != nil {
		return nil, err
	}
	return h, nil
}
