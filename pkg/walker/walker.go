package walker

// adapted from: github.com/libp2p/go-libp2p-kad-dht/crawler

import (
	"context"
	"math"
	"sync"
	"time"

	"git.d464.sh/adc/telemetry/pkg/walker/preimage"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

const ADDRESS_TTL = time.Duration(math.MaxInt64)

type Walker interface {
	Walk(ctx context.Context) error
}

func New(h host.Host, opts ...Option) (Walker, error) {
	return newImplWalker(h, opts...)
}

type walkResult struct {
	ok  *Peer
	err *Error
}

type implWalker struct {
	c *config
	h host.Host
	m *pb.ProtocolMessenger
	t *preimage.Table
}

func newImplWalker(h host.Host, opts ...Option) (*implWalker, error) {
	c := new(config)
	defaults(c)
	if err := apply(c, opts...); err != nil {
		return nil, err
	}

	messenger, err := pb.NewProtocolMessenger(&messageSender{
		h: h,
	})
	if err != nil {
		return nil, err
	}

	walker := &implWalker{
		c: c,
		h: h,
		m: messenger,
		t: preimage.Generate(),
	}

	return walker, nil
}

func (c *implWalker) Walk(ctx context.Context) error {
	cwork := make(chan peer.ID)
	cresult := make(chan walkResult, 1)
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
				res := c.walkPeer(workctx, pid)
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
			if result.ok != nil {
				c.c.observer.ObservePeer(result.ok)
				for _, bucket := range result.ok.Buckets {
					for _, addrinfo := range bucket {
						if _, ok := queried[addrinfo.ID]; !ok {
							c.h.Peerstore().AddAddrs(addrinfo.ID, addrinfo.Addrs, ADDRESS_TTL)
							pending = append(pending, addrinfo.ID)
							queried[addrinfo.ID] = struct{}{}
						}
					}
				}
			} else {
				c.c.observer.ObserveError(result.err)
			}
			inprogress -= 1
		case cworkopt <- next_id:
			pending = pending[1:]
			inprogress += 1
		}
	}

	return nil
}

func (c *implWalker) walkPeer(ctx context.Context, pid peer.ID) walkResult {
	ctx, cancel := context.WithTimeout(ctx, c.c.requestTimeout)
	defer cancel()

	walkStart := time.Now()
	walkError := &Error{
		ID:        pid,
		Addresses: c.h.Peerstore().Addrs(pid),
		Time:      walkStart,
		Err:       nil,
	}

	if err := c.h.Connect(ctx, c.h.Peerstore().PeerInfo(pid)); err != nil {
		walkError.Err = err
		return walkResult{err: walkError}
	}
	defer func() { _ = c.h.Network().ClosePeer(pid) }()

	requests := make([]Request, 0)
	buckets := make([][]peer.AddrInfo, 0)
	var addrset map[peer.ID]peer.AddrInfo = make(map[peer.ID]peer.AddrInfo)
	for _, target := range c.t.GetIDsForPeer(pid) {
		requestStart := time.Now()
		addrs, err := c.m.GetClosestPeers(ctx, pid, target)
		requestDuration := time.Since(requestStart)

		if err != nil {
			walkError.Err = err
			return walkResult{err: walkError}
		}

		requests = append(requests, Request{
			Start:    requestStart,
			Duration: requestDuration,
		})

		prevlen := len(addrset)
		for _, addr := range addrs {
			addrset[addr.ID] = *addr
		}

		if len(addrset) == prevlen {
			break
		}

		bucket := make([]peer.AddrInfo, 0, len(addrs))
		for _, a := range addrs {
			bucket = append(bucket, *a)
		}
		buckets = append(buckets, bucket)

		select {
		case <-ctx.Done():
			walkError.Err = ctx.Err()
			return walkResult{err: walkError}
		default:
		}
	}

	agent, err := c.h.Peerstore().Get(pid, "AgentVersion")
	if err != nil {
		agent = ""
	}
	protocols, err := c.h.Peerstore().GetProtocols(pid)
	if err != nil {
		protocols = []string{}
	}
	walkOk := &Peer{
		ID:        pid,
		Addresses: c.h.Peerstore().Addrs(pid),
		Agent:     agent.(string),
		Protocols: protocols,
		Buckets:   buckets,
		Requests:  requests,
	}
	return walkResult{ok: walkOk}
}
