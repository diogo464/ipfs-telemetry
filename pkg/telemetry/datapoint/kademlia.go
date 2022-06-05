package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Datapoint = (*Kademlia)(nil)
var _ Datapoint = (*KademliaQuery)(nil)

const KademliaName = "kademlia"
const KademliaQueryName = "kademliaquery"
const KademliaHandlerName = "kademliahandler"

type KademliaMessageType = pb.KademliaMessageType

const (
	KademliaMessageTypePutValue     = pb.KademliaMessageType_PUT_VALUE
	KademliaMessageTypeGetValue     = pb.KademliaMessageType_GET_VALUE
	KademliaMessageTypeAddProvider  = pb.KademliaMessageType_ADD_PROVIDER
	KademliaMessageTypeGetProviders = pb.KademliaMessageType_GET_PROVIDERS
	KademliaMessageTypeFindNode     = pb.KademliaMessageType_FIND_NODE
	KademliaMessageTypePing         = pb.KademliaMessageType_PING
)

var KademliaMessageTypes = []KademliaMessageType{
	KademliaMessageTypePutValue,
	KademliaMessageTypeGetValue,
	KademliaMessageTypeAddProvider,
	KademliaMessageTypeGetProviders,
	KademliaMessageTypeFindNode,
	KademliaMessageTypePing,
}

var KademliaMessageTypeString = map[KademliaMessageType]string{
	KademliaMessageTypePutValue:     "putvalue",
	KademliaMessageTypeGetValue:     "getvalue",
	KademliaMessageTypeAddProvider:  "addprovider",
	KademliaMessageTypeGetProviders: "getproviders",
	KademliaMessageTypeFindNode:     "findnode",
	KademliaMessageTypePing:         "ping",
}

type Kademlia struct {
	Timestamp   time.Time                      `json:"timestamp"`
	MessagesIn  map[KademliaMessageType]uint64 `json:"messages_in"`
	MessagesOut map[KademliaMessageType]uint64 `json:"messages_out"`
}

func (*Kademlia) sealed()                   {}
func (*Kademlia) GetName() string           { return KademliaName }
func (p *Kademlia) GetTimestamp() time.Time { return p.Timestamp }
func (p *Kademlia) GetSizeEstimate() uint32 {
	return estimateTimestampSize + uint32(len(p.MessagesIn))*(4+8) + uint32(len(p.MessagesOut))*(4+8)
}
func (p *Kademlia) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Kademlia{
			Kademlia: KademliaToPB(p),
		},
	}
}

func KademliaFromPB(in *pb.Kademlia) (*Kademlia, error) {
	return &Kademlia{
		Timestamp:   pbutils.TimeFromPB(in.GetTimestamp()),
		MessagesIn:  kademliaCountMapFromPB(in.GetMessagesIn()),
		MessagesOut: kademliaCountMapFromPB(in.GetMessagesOut()),
	}, nil
}

func KademliaToPB(in *Kademlia) *pb.Kademlia {
	return &pb.Kademlia{
		Timestamp:   pbutils.TimeToPB(&in.Timestamp),
		MessagesIn:  kademliaCountMapToPB(in.MessagesIn),
		MessagesOut: kademliaCountMapToPB(in.MessagesOut),
	}
}

type KademliaQuery struct {
	Timestamp time.Time           `json:"timestamp"`
	Peer      peer.ID             `json:"peer"`
	QueryType KademliaMessageType `json:"query_type"`
	Duration  time.Duration       `json:"duration"`
}

func (*KademliaQuery) sealed()                   {}
func (*KademliaQuery) GetName() string           { return KademliaQueryName }
func (p *KademliaQuery) GetTimestamp() time.Time { return p.Timestamp }
func (p *KademliaQuery) GetSizeEstimate() uint32 {
	return estimateTimestampSize + estimatePeerIdSize + 4 + estimateDurationSize
}
func (p *KademliaQuery) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_KademliaQuery{
			KademliaQuery: KademliaQueryToPB(p),
		},
	}
}

func KademliaQueryFromPB(in *pb.KademliaQuery) (*KademliaQuery, error) {
	p, err := peer.Decode(in.GetPeer())
	if err != nil {
		return nil, err
	}
	return &KademliaQuery{
		Timestamp: pbutils.TimeFromPB(in.GetTimestamp()),
		Peer:      p,
		QueryType: in.GetQueryType(),
		Duration:  pbutils.DurationFromPB(in.GetDuration()),
	}, nil
}

func KademliaQueryToPB(p *KademliaQuery) *pb.KademliaQuery {
	return &pb.KademliaQuery{
		Timestamp: pbutils.TimeToPB(&p.Timestamp),
		Peer:      p.Peer.Pretty(),
		QueryType: p.QueryType,
		Duration:  pbutils.DurationToPB(&p.Duration),
	}
}

type KademliaHandler struct {
	Timestamp       time.Time           `json:"timestamp"`
	HandlerType     KademliaMessageType `json:"handler_type"`
	HandlerDuration time.Duration       `json:"handler_duration"`
	WriteDuration   time.Duration       `json:"write_duration"`
}

func (*KademliaHandler) sealed()                   {}
func (*KademliaHandler) GetName() string           { return KademliaHandlerName }
func (p *KademliaHandler) GetTimestamp() time.Time { return p.Timestamp }
func (p *KademliaHandler) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 4 + 2*estimateDurationSize
}
func (p *KademliaHandler) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_KademliaHandler{
			KademliaHandler: KademliaHandlerToPB(p),
		},
	}
}

func KademliaHandlerFromPB(in *pb.KademliaHandler) (*KademliaHandler, error) {
	return &KademliaHandler{
		Timestamp:       pbutils.TimeFromPB(in.GetTimestamp()),
		HandlerType:     in.GetHandlerType(),
		HandlerDuration: pbutils.DurationFromPB(in.GetHandlerDuration()),
		WriteDuration:   pbutils.DurationFromPB(in.GetWriteDuration()),
	}, nil
}

func KademliaHandlerToPB(p *KademliaHandler) *pb.KademliaHandler {
	return &pb.KademliaHandler{
		Timestamp:       pbutils.TimeToPB(&p.Timestamp),
		HandlerType:     p.HandlerType,
		HandlerDuration: pbutils.DurationToPB(&p.HandlerDuration),
		WriteDuration:   pbutils.DurationToPB(&p.WriteDuration),
	}
}

func kademliaCountMapFromPB(in map[uint32]uint64) map[KademliaMessageType]uint64 {
	out := make(map[KademliaMessageType]uint64, len(in))
	// ignore unkown message types
	for k, v := range in {
		switch k {
		case uint32(KademliaMessageTypePutValue):
			out[KademliaMessageTypePutValue] = v
		case uint32(KademliaMessageTypeGetValue):
			out[KademliaMessageTypeGetValue] = v
		case uint32(KademliaMessageTypeAddProvider):
			out[KademliaMessageTypeAddProvider] = v
		case uint32(KademliaMessageTypeGetProviders):
			out[KademliaMessageTypeGetProviders] = v
		case uint32(KademliaMessageTypeFindNode):
			out[KademliaMessageTypeFindNode] = v
		case uint32(KademliaMessageTypePing):
			out[KademliaMessageTypePing] = v
		default:
		}
	}
	return out
}

func kademliaCountMapToPB(in map[KademliaMessageType]uint64) map[uint32]uint64 {
	out := make(map[uint32]uint64, len(in))
	for k, v := range in {
		out[uint32(k)] = v
	}
	return out
}
