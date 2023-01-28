package crawler

import (
	"context"
	"sync"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/crawler/metrics"
	"github.com/diogo464/telemetry/internal/utils"
	"github.com/diogo464/telemetry/walker"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var _ (walker.Observer) = (*Crawler)(nil)

type counters struct {
	peers  *atomic.Uint64
	tpeers *atomic.Uint64
	errors *atomic.Uint64
}

func newCounters() *counters {
	return &counters{
		peers:  atomic.NewUint64(0),
		tpeers: atomic.NewUint64(0),
		errors: atomic.NewUint64(0),
	}
}

func (c *counters) clone() *counters {
	return &counters{
		peers:  atomic.NewUint64(c.peers.Load()),
		tpeers: atomic.NewUint64(c.tpeers.Load()),
		errors: atomic.NewUint64(c.errors.Load()),
	}
}

type Crawler struct {
	l    *zap.Logger
	h    host.Host
	w    walker.Walker
	opts *options

	peers_mu sync.Mutex
	peers    map[peer.ID]struct{} // active peers
	tpeers   map[peer.ID]struct{} //active telemetry peers

	completed *atomic.Uint64
	cnow      *counters
	cold      *counters
}

func NewCrawler(h host.Host, o ...Option) (*Crawler, error) {
	opts := defaults()
	if err := apply(opts, o...); err != nil {
		return nil, err
	}

	c := &Crawler{
		l:    opts.logger,
		h:    h,
		w:    nil,
		opts: opts,

		peers:  make(map[peer.ID]struct{}),
		tpeers: make(map[peer.ID]struct{}),

		completed: atomic.NewUint64(0),
		cnow:      newCounters(),
		cold:      newCounters(),
	}

	walkerOpts := []walker.Option{}
	walkerOpts = append(walkerOpts, opts.walkerOpts...)
	walkerOpts = append(walkerOpts, walker.WithObserver(c))

	w, err := walker.New(h, walkerOpts...)
	if err != nil {
		return nil, err
	}
	c.w = w

	m, err := metrics.New(opts.meterProvider)
	if err != nil {
		return nil, err
	}
	m.RegisterCallback(func(ctx context.Context) {
		m.PeersCurrentCrawl.Observe(ctx, int64(c.cnow.peers.Load()))
		m.PeersTelemetryCurrentCrawl.Observe(ctx, int64(c.cnow.tpeers.Load()))
		m.ErrorsCurrentCrawl.Observe(ctx, int64(c.cnow.errors.Load()))

		m.PeersLastCrawl.Observe(ctx, int64(c.cold.peers.Load()))
		m.PeersTelemetryLastCrawl.Observe(ctx, int64(c.cold.tpeers.Load()))
		m.ErrorsLastCrawl.Observe(ctx, int64(c.cold.errors.Load()))

		m.CompletedCrawls.Observe(ctx, int64(c.completed.Load()))
	})

	return c, nil
}

func (c *Crawler) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := c.w.Walk(ctx); err != nil {
			return err
		}

		c.cold = c.cnow.clone()
		c.cnow = newCounters()
		c.completed.Inc()

		c.peers = make(map[peer.ID]struct{})
		c.tpeers = make(map[peer.ID]struct{})
	}
}

// ObservePeer implements walker.Observer
func (c *Crawler) ObservePeer(p *walker.Peer) {
	hasTelemetry := utils.SliceAny(p.Protocols, func(t string) bool { return t == telemetry.ID_TELEMETRY })

	c.peers_mu.Lock()
	{
		if _, ok := c.peers[p.ID]; !ok {
			c.cnow.peers.Inc()
			if hasTelemetry {
				c.cnow.tpeers.Inc()
			}
		}
		c.peers[p.ID] = struct{}{}
		if hasTelemetry {
			c.tpeers[p.ID] = struct{}{}
		}
	}
	c.peers_mu.Unlock()

	if hasTelemetry {
		c.opts.observer.ObservePeer(p)
		c.l.Info("found telemetry peer", zap.String("peer", p.ID.Pretty()))
	}
}

// ObserveError implements walker.Observer
func (c *Crawler) ObserveError(e *walker.Error) {
	c.cnow.errors.Inc()
	c.l.Info("error", zap.Any("error", e))
}
