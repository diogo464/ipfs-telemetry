package crawler

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type DiscoveryNotification struct {
	ID        peer.ID               `json:"id"`
	Addresses []multiaddr.Multiaddr `json:"addresses"`
}
