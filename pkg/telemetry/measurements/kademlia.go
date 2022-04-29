package measurements

import (
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
)

var kademlia Kademlia = nil

type KademliaQueryTypeKey struct{}

type Kademlia interface {
	IncMessageIn(datapoint.KademliaMessageType)
	IncMessageOut(datapoint.KademliaMessageType)
	PushQuery(peer.ID, datapoint.KademliaMessageType, time.Duration)
	PushHandler(p peer.ID, m datapoint.KademliaMessageType, handler time.Duration, write time.Duration)
}

func ConvertKademliaMessageType(in pb.Message_MessageType) datapoint.KademliaMessageType {
	switch in {
	case pb.Message_PUT_VALUE:
		return datapoint.KademliaMessageTypePutValue
	case pb.Message_GET_VALUE:
		return datapoint.KademliaMessageTypeGetValue
	case pb.Message_ADD_PROVIDER:
		return datapoint.KademliaMessageTypeAddProvider
	case pb.Message_GET_PROVIDERS:
		return datapoint.KademliaMessageTypeGetProviders
	case pb.Message_FIND_NODE:
		return datapoint.KademliaMessageTypeFindNode
	case pb.Message_PING:
		return datapoint.KademliaMessageTypePing
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
