package measurements

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

var kademlia Kademlia = nil

type KademliaQueryTypeKey struct{}

type Kademlia interface {
	IncMessageIn(snapshot.KademliaMessageType)
	IncMessageOut(snapshot.KademliaMessageType)
	PushQuery(peer.ID, snapshot.KademliaMessageType, time.Duration)
	PushHandler(p peer.ID, m snapshot.KademliaMessageType, handler time.Duration, write time.Duration)
}

func ConvertKademliaMessageType(in pb.Message_MessageType) snapshot.KademliaMessageType {
	switch in {
	case pb.Message_PUT_VALUE:
		return snapshot.KademliaMessageTypePutValue
	case pb.Message_GET_VALUE:
		return snapshot.KademliaMessageTypeGetValue
	case pb.Message_ADD_PROVIDER:
		return snapshot.KademliaMessageTypeAddProvider
	case pb.Message_GET_PROVIDERS:
		return snapshot.KademliaMessageTypeGetProviders
	case pb.Message_FIND_NODE:
		return snapshot.KademliaMessageTypeFindNode
	case pb.Message_PING:
		return snapshot.KademliaMessageTypePing
	default:
		panic("unimplemented")
	}
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
