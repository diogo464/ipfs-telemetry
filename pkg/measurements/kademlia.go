package measurements

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
)

var kademlia Kademlia = nil

type Kademlia interface {
	PushQueryTiming(peer.ID, time.Duration)
}

func KademliaRegister(k Kademlia) {
	if kademlia != nil {
		panic("should not happend")
	}
	kademlia = k
}

func WithKademlia(fn func(k Kademlia)) {
	if kademlia != nil {
		fn(kademlia)
	}
}
