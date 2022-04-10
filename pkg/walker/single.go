package walker

import (
	"context"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

// Return all the peers in p's buckets
func Dump(ctx context.Context, h host.Host, p peer.AddrInfo) ([]peer.AddrInfo, error) {
	walker, err := newImplWalker(h, WithConcurrency(1))
	if err != nil {
		return nil, err
	}

	h.Peerstore().AddAddrs(p.ID, p.Addrs, ADDRESS_TTL)
	result := walker.walkPeer(ctx, p.ID)
	if result.ok != nil {
		addrs := make([]peer.AddrInfo, 0)
		for _, bucket := range result.ok.Buckets {
			addrs = append(addrs, bucket...)
		}
		return addrs, nil
	} else {
		return nil, result.err.Err
	}
}
