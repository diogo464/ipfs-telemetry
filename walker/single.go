package walker

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

// Return all the peers in p's buckets
func Dump(ctx context.Context, h host.Host, p peer.AddrInfo) ([]BucketEntry, error) {
	walker, err := newImplWalker(h, WithConcurrency(1))
	if err != nil {
		return nil, err
	}

	h.Peerstore().AddAddrs(p.ID, p.Addrs, peerstore.PermanentAddrTTL)
	result := walker.walkPeerTask(ctx, p.ID)
	if result.ok != nil {
		return result.ok.Buckets, nil
	} else {
		return nil, result.err.Err
	}
}
