package walker

// adapted from: github.com/libp2p/go-libp2p-kad-dht/crawler

import (
	"context"
	"sync"
	"time"

	"github.com/diogo464/telemetry/vecdeque"
	"github.com/diogo464/telemetry/walker/preimage"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

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
	h         host.Host
	opts      *options
	messenger *pb.ProtocolMessenger
	table     *preimage.Table
	wg        sync.WaitGroup
}

func newImplWalker(h host.Host, opts ...Option) (*implWalker, error) {
	c := new(options)
	defaults(c)
	if err := apply(c, opts...); err != nil {
		return nil, err
	}

	messenger, err := pb.NewProtocolMessenger(&MessageSender{
		h: h,
	})
	if err != nil {
		return nil, err
	}

	walker := &implWalker{
		h:         h,
		opts:      c,
		messenger: messenger,
		table:     preimage.Generate(),
	}

	return walker, nil
}

func (c *implWalker) Walk(ctx context.Context) error {
	cwork := make(chan peer.ID)
	cresult := make(chan walkResult, 1)
	defer close(cwork)

	var err error = nil
	inprogress := 0
	pending := vecdeque.New[peer.ID]()
	queried := make(map[peer.ID]struct{})
	interval := time.NewTicker(c.opts.interval)

	for _, addr := range c.opts.seeds {
		c.h.Peerstore().AddAddrs(addr.ID, addr.Addrs, peerstore.PermanentAddrTTL)
		pending.PushBack(addr.ID)
	}

LOOP:
	for pending.Len() > 0 || inprogress > 0 {
		var intervalChan <-chan time.Time
		if pending.Len() > 0 && inprogress < int(c.opts.concurrency) {
			intervalChan = interval.C
		}

		select {
		case result := <-cresult:
			if result.ok != nil {
				c.opts.observer.ObservePeer(result.ok)
				for _, addrinfo := range result.ok.Buckets {
					if _, ok := queried[addrinfo.ID]; !ok {
						c.h.Peerstore().AddAddrs(addrinfo.ID, addrinfo.Addrs, peerstore.PermanentAddrTTL)
						queried[addrinfo.ID] = struct{}{}
						pending.PushBack(addrinfo.ID)
					}
				}
			} else {
				c.opts.observer.ObserveError(result.err)
			}
			inprogress -= 1
		case <-intervalChan:
			inprogress += 1
			pid := pending.PopFront()
			c.walkPeer(ctx, cresult, pid)
		case <-ctx.Done():
			err = ctx.Err()
			break LOOP
		}
	}

	c.wg.Wait()

	return err
}

func (c *implWalker) walkPeerTask(ctx context.Context, pid peer.ID) walkResult {
	connCtx, connCancel := context.WithTimeout(ctx, c.opts.connectTimeout)
	defer connCancel()

	walkStart := time.Now()
	walkError := &Error{
		ID:        pid,
		Addresses: c.h.Peerstore().Addrs(pid),
		Time:      walkStart,
		Err:       nil,
	}

	connectStart := time.Now()
	if err := c.h.Connect(connCtx, c.h.Peerstore().PeerInfo(pid)); err != nil {
		walkError.Err = err
		return walkResult{err: walkError}
	}
	connectDuration := time.Since(connectStart)
	defer func() { _ = c.h.Network().ClosePeer(pid) }()

	reqCtx, reqCancel := context.WithTimeout(ctx, c.opts.requestTimeout)
	defer reqCancel()

	requests := make([]Request, 0)
	buckets := make([]BucketEntry, 0)
	var addrset map[peer.ID]peer.AddrInfo = make(map[peer.ID]peer.AddrInfo)
	for _, target := range c.table.GetIDsForPeer(pid) {
		requestStart := time.Now()
		addrs, err := c.messenger.GetClosestPeers(reqCtx, pid, target)
		requestDuration := time.Since(requestStart)

		if err != nil {
			walkError.Err = err
			return walkResult{err: walkError}
		}

		requests = append(requests, Request{
			Start:    requestStart,
			Duration: requestDuration,
		})

		for _, addr := range addrs {
			if _, ok := addrset[addr.ID]; !ok {
				buckets = append(buckets, BucketEntry(*addr))
			}
		}

		prevlen := len(addrset)
		for _, addr := range addrs {
			addrset[addr.ID] = *addr
		}

		if len(addrset) == prevlen {
			break
		}

		select {
		case <-reqCtx.Done():
			walkError.Err = reqCtx.Err()
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
		ID:              pid,
		Addresses:       c.h.Peerstore().Addrs(pid),
		Agent:           agent.(string),
		Protocols:       protocols,
		Buckets:         buckets,
		Requests:        requests,
		ConnectStart:    connectStart,
		ConnectDuration: connectDuration,
	}
	return walkResult{ok: walkOk}
}

func (c *implWalker) walkPeer(ctx context.Context, cresult chan<- walkResult, pid peer.ID) {
	c.wg.Add(1)
	go func() {
		res := c.walkPeerTask(ctx, pid)
		select {
		case cresult <- res:
		case <-ctx.Done():
		}
		c.wg.Add(-1)
	}()
}
