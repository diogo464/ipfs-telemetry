package crawler

import (
	"context"
	"sync"

	pb "git.d464.sh/adc/telemetry/pkg/proto/crawler"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"git.d464.sh/adc/telemetry/pkg/walker"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/emptypb"
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

		subscribers: make(map[chan<- peer.ID]struct{}),
	}

	w, err := walker.New(h, walker.WithObserver(walker.NewMultiObserver(c, opts.observer)))
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
		CompletedCrawls.Inc()

		c.peers = make(map[peer.ID]struct{})
		c.tpeers = make(map[peer.ID]struct{})
	}
}

// ObservePeer implements walker.Observer
func (c *Crawler) ObservePeer(p *walker.Peer) {
	hasTelemetry, err := c.peerHasTelemetry(p.ID)
	if err != nil { // dont stop the crawl, just ignore this peer
		return
	}

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
func (*Crawler) ObserveError(*walker.Error) {
	ErrorsCurrentCrawl.Inc()
}

func (c *Crawler) Subscribe(req *emptypb.Empty, stream pb.Crawler_SubscribeServer) error {
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

func (c *Crawler) peerHasTelemetry(p peer.ID) (bool, error) {
	protocols, err := c.h.Peerstore().GetProtocols(p)
	if err != nil {
		return false, err
	}

	for _, protocol := range protocols {
		if protocol == telemetry.ID_TELEMETRY {
			return true, nil
		}
	}
	return false, nil
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
