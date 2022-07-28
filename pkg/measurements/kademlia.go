package measurements

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

var kademlia []Kademlia = []Kademlia{}

type KademliaQueryTypeKey struct{}

type Kademlia interface {
	IncMessageIn(pb.Message_MessageType)
	IncMessageOut(pb.Message_MessageType)
	PushQuery(peer.ID, pb.Message_MessageType, time.Duration)
	PushHandler(p peer.ID, m pb.Message_MessageType, handler time.Duration, write time.Duration)
}

func KademliaRegister(k Kademlia) {
	for _, kad := range kademlia {
		if kad == k {
			return
		}
	}
	kademlia = append(kademlia, k)
}

func WithKademlia(fn func(k Kademlia)) {
	for _, kad := range kademlia {
		fn(kad)
	}
}
