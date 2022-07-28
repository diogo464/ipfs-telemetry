package collectors

import (
	"context"
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/ipfs_telemetry/pkg/measurements"
	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p-core/peer"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	"go.uber.org/atomic"
)

var _ telemetry.Collector = (*kademliaCollector)(nil)
var _ measurements.Kademlia = (*kademliaCollector)(nil)

type kademliaCollector struct {
	messages_in  map[datapoint.KademliaMessageType]*atomic.Uint64
	messages_out map[datapoint.KademliaMessageType]*atomic.Uint64
	kadpb        datapoint.Kademlia
}

func Kademlia() telemetry.Collector {
	c := &kademliaCollector{
		messages_in:  make(map[datapoint.KademliaMessageType]*atomic.Uint64),
		messages_out: make(map[datapoint.KademliaMessageType]*atomic.Uint64),
		kadpb: datapoint.Kademlia{
			MessagesIn:  make(map[datapoint.KademliaMessageType]uint64),
			MessagesOut: make(map[datapoint.KademliaMessageType]uint64),
		},
	}
	for _, ty := range datapoint.KademliaMessageTypes {
		c.messages_in[ty] = atomic.NewUint64(0)
		c.messages_out[ty] = atomic.NewUint64(0)
	}
	measurements.KademliaRegister(c)
	return c
}

// IncMessageIn implements measurements.Kademlia
func (c *kademliaCollector) IncMessageIn(pbty pb.Message_MessageType) {
	ty := ConvertKademliaMessageType(pbty)
	c.messages_in[ty].Inc()
}

// IncMessageOut implements measurements.Kademlia
func (c *kademliaCollector) IncMessageOut(pbty pb.Message_MessageType) {
	ty := ConvertKademliaMessageType(pbty)
	c.messages_out[ty].Inc()
}

// PushHandler implements measurements.Kademlia
func (*kademliaCollector) PushHandler(p peer.ID, m pb.Message_MessageType, handler time.Duration, write time.Duration) {
}

// PushQuery implements measurements.Kademlia
func (*kademliaCollector) PushQuery(peer.ID, pb.Message_MessageType, time.Duration) {
}

// Descriptor implements telemetry.Collector
func (*kademliaCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.KademliaName,
	}
}

// Open implements telemetry.Collector
func (*kademliaCollector) Open() {
}

// Close implements telemetry.Collector
func (*kademliaCollector) Close() {
}

// Collect implements telemetry.Collector
func (c *kademliaCollector) Collect(_ context.Context, stream *telemetry.Stream) error {
	for k, v := range c.messages_in {
		c.kadpb.MessagesIn[k] = v.Load()
	}
	for k, v := range c.messages_out {
		c.kadpb.MessagesOut[k] = v.Load()
	}
	c.kadpb.Timestamp = datapoint.NewTimestamp()
	return datapoint.KademliaSerialize(&c.kadpb, stream)
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
