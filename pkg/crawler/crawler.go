package crawler

// adapted from: github.com/libp2p/go-libp2p-kad-dht/crawler

import (
	"context"
	"math"
	"sync"
	"time"

	"git.d464.sh/adc/telemetry/pkg/preimage"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

const ADDRESS_TTL = time.Duration(math.MaxInt64)

type Crawler interface {
	Crawl(ctx context.Context) error
}

func NewCrawler(h host.Host, handler EventHandler, opts ...Option) (Crawler, error) {
	return newImplCrawler(h, handler, opts...)
}

type crawlResult struct {
	peer    peer.ID
	peers   []peer.AddrInfo
	err     error
	usererr error
}

type implCrawler struct {
	c       *config
	h       host.Host
	m       *pb.ProtocolMessenger
	t       *preimage.Table
	handler EventHandler
}

func newImplCrawler(h host.Host, handler EventHandler, opts ...Option) (*implCrawler, error) {
	c := new(config)
	defaults(c)
	if err := apply(c, opts...); err != nil {
		return nil, err
	}

	messenger, err := pb.NewProtocolMessenger(&scraperMessageSender{
		h: h,
	})
	if err != nil {
		return nil, err
	}

	crawler := &implCrawler{
		c:       c,
		h:       h,
		m:       messenger,
		t:       preimage.Generate(),
		handler: handler,
	}

	return crawler, nil
}

func (c *implCrawler) Crawl(ctx context.Context) error {
	handler := c.handler
	cwork := make(chan peer.ID)
	cresult := make(chan crawlResult, 1)
	workctx, cancel := context.WithCancel(ctx)
	wg := new(sync.WaitGroup)

	defer wg.Wait()
	defer close(cwork)
	defer cancel()

	wg.Add(int(c.c.concurrency))
	for i := 0; i < int(c.c.concurrency); i++ {
		go func() {
			defer wg.Done()
		LOOP:
			for pid := range cwork {
				res := c.crawlPeer(workctx, handler, pid)
				select {
				case cresult <- res:
				case <-workctx.Done():
					break LOOP
				}
			}
		}()
	}

	inprogress := 0
	pending := make([]peer.ID, 0)
	queried := make(map[peer.ID]struct{})

	for _, addr := range c.c.seeds {
		c.h.Peerstore().AddAddrs(addr.ID, addr.Addrs, ADDRESS_TTL)
		pending = append(pending, addr.ID)
	}

	for len(pending) > 0 || inprogress > 0 {
		var next_id peer.ID
		var cworkopt chan peer.ID

		if len(pending) > 0 {
			next_id = pending[0]
			cworkopt = cwork
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		select {
		case result := <-cresult:
			if result.usererr != nil {
				return result.usererr
			}
			if result.err == nil {
				for _, addr := range result.peers {
					if _, ok := queried[addr.ID]; !ok {
						c.h.Peerstore().AddAddrs(addr.ID, addr.Addrs, ADDRESS_TTL)
						pending = append(pending, addr.ID)
						queried[addr.ID] = struct{}{}
					}
				}
			}
			inprogress -= 1
		case cworkopt <- next_id:
			pending = pending[1:]
			inprogress += 1
		}
	}

	return nil
}

func (c *implCrawler) crawlPeer(ctx context.Context, handler EventHandler, pid peer.ID) crawlResult {
	ctx, cancel := context.WithTimeout(ctx, c.c.requestTimeout)
	defer cancel()

	if err := c.h.Connect(ctx, c.h.Peerstore().PeerInfo(pid)); err != nil {
		return crawlResult{err: err}
	}
	defer func() { _ = c.h.Network().ClosePeer(pid) }()

	if err := handler.OnConnect(pid); err != nil {
		return crawlResult{usererr: err}
	}

	var result crawlResult
	var addrset map[peer.ID]peer.AddrInfo = make(map[peer.ID]peer.AddrInfo)
	for _, target := range c.t.GetIDsForPeer(pid) {
		addrs, err := c.m.GetClosestPeers(ctx, pid, target)
		if err != nil {
			result.err = err
			break
		}

		prevlen := len(addrset)
		for _, addr := range addrs {
			addrset[addr.ID] = *addr
		}

		if len(addrset) == prevlen {
			break
		}

		select {
		case <-ctx.Done():
			return crawlResult{err: ctx.Err()}
		default:
		}
	}

	discovered := make([]peer.AddrInfo, 0, len(addrset))
	for _, v := range addrset {
		discovered = append(discovered, v)
	}
	result.peers = discovered

	if result.err == nil {
		result.usererr = handler.OnFinish(pid, discovered)
	} else {
		result.usererr = handler.OnFail(pid, result.err)
	}

	return result
}
