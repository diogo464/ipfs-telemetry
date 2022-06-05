package crawler

import (
	"context"
	"sync"

	pb "github.com/diogo464/telemetry/pkg/proto/crawler"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/utils"
	"github.com/diogo464/telemetry/pkg/walker"
	"github.com/gogo/protobuf/types"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"go.uber.org/atomic"
)

var _ (walker.Observer) = (*Crawler)(nil)

const DEFAULT_CHANNEL_BUFFER_SIZE = 32

type Crawler struct {
	pb.UnimplementedCrawlerServer
	h    host.Host
	w    walker.Walker
	opts *options

	peers_mu sync.Mutex
	peers    map[peer.ID]struct{} // active peers
	tpeers   map[peer.ID]struct{} //active telemetry peers
	errors   *atomic.Uint32

	subscribers_mu sync.Mutex
	subscribers    map[chan<- peer.ID]struct{}
}

func NewCrawler(h host.Host, o ...Option) (*Crawler, error) {
	opts := defaults()
	if err := apply(opts, o...); err != nil {
		return nil, err
	}

	c := &Crawler{
		h:    h,
		w:    nil,
		opts: opts,

		peers:  make(map[peer.ID]struct{}),
		tpeers: make(map[peer.ID]struct{}),
		errors: atomic.NewUint32(0),

		subscribers: make(map[chan<- peer.ID]struct{}),
	}

	w, err := walker.New(h, walker.WithObserver(walker.NewMultiObserver(c, opts.observer)), walker.WithConcurrency(uint(opts.concurrency)))
	if err != nil {
		return nil, err
	}

	c.w = w
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

		PeersLastCrawl.Set(float64(len(c.peers)))
		PeersTelemetryLastCrawl.Set(float64(len(c.tpeers)))
		PeersCurrentCrawl.Set(0)
		PeersTelemetryCurrentCrawl.Set(0)
		ErrorsLastCrawl.Set(float64(c.errors.Load()))
		ErrorsCurrentCrawl.Set(0)
		c.errors.Store(0)
		CompletedCrawls.Inc()

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
			PeersCurrentCrawl.Inc()
			if hasTelemetry {
				PeersTelemetryCurrentCrawl.Inc()
			}
		}
		c.peers[p.ID] = struct{}{}
		if hasTelemetry {
			c.tpeers[p.ID] = struct{}{}
		}
	}
	c.peers_mu.Unlock()

	if hasTelemetry {
		c.broadcastPeer(p.ID)
	}
}

// ObserveError implements walker.Observer
func (c *Crawler) ObserveError(*walker.Error) {
	ErrorsCurrentCrawl.Inc()
	c.errors.Inc()
}

func (c *Crawler) Subscribe(req *types.Empty, stream pb.Crawler_SubscribeServer) error {
	csubscribe := make(chan peer.ID, DEFAULT_CHANNEL_BUFFER_SIZE)
	c.subscribers_mu.Lock()
	c.subscribers[csubscribe] = struct{}{}
	c.subscribers_mu.Unlock()
	defer func() {
		c.subscribers_mu.Lock()
		delete(c.subscribers, csubscribe)
		c.subscribers_mu.Unlock()
	}()

	go func() {
		for _, p := range c.cloneKnownTelemetryPeers() {
			csubscribe <- p
		}
	}()

	for p := range csubscribe {
		err := stream.Send(&pb.SubscribeItem{
			PeerId: p.Pretty(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Crawler) broadcastPeer(p peer.ID) {
	c.subscribers_mu.Lock()
	defer c.subscribers_mu.Unlock()
	for subscriber := range c.subscribers {
		select {
		case subscriber <- p:
		default:
		}
	}
}

func (c *Crawler) cloneKnownTelemetryPeers() []peer.ID {
	c.peers_mu.Lock()
	defer c.peers_mu.Unlock()
	peers := make([]peer.ID, 0, len(c.tpeers))
	for k := range c.tpeers {
		peers = append(peers, k)
	}
	return peers
}
