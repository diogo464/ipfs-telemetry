package crawler

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

func CrawSingle(ctx context.Context, h host.Host, p peer.AddrInfo) ([]peer.AddrInfo, error) {
	crawler, err := newImplCrawler(h, NullEventHandler{}, WithConcurrency(1))
	if err != nil {
		return nil, err
	}

	h.Peerstore().AddAddrs(p.ID, p.Addrs, ADDRESS_TTL)
	result := crawler.crawlPeer(ctx, crawler.handler, p.ID)

	return result.peers, result.err
}
