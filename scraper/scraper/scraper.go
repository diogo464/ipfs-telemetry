// https://github.com/libp2p/go-libp2p-kad-dht/blob/master/crawler/crawler.go
package scraper

import (
	"context"
	"math"
	"sync"
	"time"

	"d464.sh/scraper/preimage"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"github.com/libp2p/go-msgio/protoio"
)

const gQUERY_TIMEOUT = time.Second * 25

type EventHandler interface {
	ScrapeBegin(p peer.ID)
	ScrapeFailed(p peer.ID, err error)
	ScrapeFinished(p peer.ID, addrs []peer.AddrInfo)
}

type scraperMessageSender struct {
	h host.Host
}

func (ms *scraperMessageSender) SendRequest(ctx context.Context, p peer.ID, pmes *pb.Message) (*pb.Message, error) {
	stream, err := ms.h.NewStream(ctx, p, dht.DefaultProtocols...)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	writer := protoio.NewDelimitedWriter(stream)
	if err := writer.WriteMsg(pmes); err != nil {
		return nil, err
	}

	msg := new(pb.Message)
	reader := protoio.NewDelimitedReader(stream, network.MessageSizeMax)
	if err := reader.ReadMsg(msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (ms *scraperMessageSender) SendMessage(ctx context.Context, p peer.ID, pmes *pb.Message) error {
	stream, err := ms.h.NewStream(ctx, p, dht.DefaultProtocols...)
	if err != nil {
		return err
	}
	defer stream.Close()

	writer := protoio.NewDelimitedWriter(stream)
	if err := writer.WriteMsg(pmes); err != nil {
		return err
	}

	return nil
}

type scraperResult struct {
	peer  peer.ID
	peers []peer.AddrInfo
	err   error
}

type Scraper struct {
	h         host.Host
	table     *preimage.Table
	messenger *pb.ProtocolMessenger
}

func NewScraper(h host.Host, table *preimage.Table) (*Scraper, error) {
	messenger, err := pb.NewProtocolMessenger(&scraperMessageSender{
		h: h,
	})
	if err != nil {
		return nil, err
	}

	return &Scraper{
		h:         h,
		table:     table,
		messenger: messenger,
	}, nil
}

func (s *Scraper) Run(ctx context.Context, seeds []peer.AddrInfo, handler EventHandler, workers int) error {
	const ADDR_TTL = time.Duration(math.MaxInt64)

	cwork := make(chan peer.ID)
	cresult := make(chan scraperResult, 1)
	wg := new(sync.WaitGroup)

	defer wg.Wait()
	defer close(cwork)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for id := range cwork {
				res := s.workerScrapePeer(ctx, handler, id)
				cresult <- res
			}
		}()
	}

	inprogress := 0
	pending := make([]peer.ID, 0)
	queried := make(map[peer.ID]struct{})

	for _, addr := range seeds {
		s.h.Peerstore().AddAddrs(addr.ID, addr.Addrs, ADDR_TTL)
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
		case result := <-cresult:
			if result.err == nil {
				for _, addr := range result.peers {
					if _, ok := queried[addr.ID]; !ok {
						s.h.Peerstore().AddAddrs(addr.ID, addr.Addrs, ADDR_TTL)
						pending = append(pending, addr.ID)
						queried[addr.ID] = struct{}{}
					}
				}
			}
			inprogress -= 1
		case cworkopt <- next_id:
			pending = pending[1:]
			inprogress += 1
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (s *Scraper) workerScrapePeer(ctx context.Context, handler EventHandler, p peer.ID) scraperResult {
	ctx, cancel := context.WithTimeout(ctx, gQUERY_TIMEOUT)
	defer cancel()

	info := s.h.Peerstore().PeerInfo(p)
	if err := s.h.Connect(network.WithUseTransient(ctx, "scraper"), info); err != nil {
		return scraperResult{err: err}
	}

	defer func() {
		for _, c := range s.h.Network().ConnsToPeer(p) {
			c.Close()
		}
	}()

	handler.ScrapeBegin(p)

	var result scraperResult
	peer_addresses := make([]peer.AddrInfo, 0)
	for _, c := range s.table.GetIDsForPeer(p) {
		addrs, err := s.messenger.GetClosestPeers(ctx, p, c)
		if err != nil {
			result.err = err
			break
		}

		for _, addr := range addrs {
			peer_addresses = append(peer_addresses, *addr)
		}
	}

	if result.err == nil {
		handler.ScrapeFinished(p, peer_addresses)
	} else {
		handler.ScrapeFailed(p, result.err)
	}

	result.peer = p
	result.peers = peer_addresses
	return result
}
