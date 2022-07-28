package datapoint

import (
	"time"

	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	"github.com/libp2p/go-libp2p-core/peer"
)

const KademliaName = "kademlia"
const KademliaQueryName = "kademliaquery"
const KademliaHandlerName = "kademliahandler"

type KademliaMessageType = uint32

const (
	KademliaMessageTypePutValue     = uint32(pb.KademliaMessageType_PUT_VALUE)
	KademliaMessageTypeGetValue     = uint32(pb.KademliaMessageType_GET_VALUE)
	KademliaMessageTypeAddProvider  = uint32(pb.KademliaMessageType_ADD_PROVIDER)
	KademliaMessageTypeGetProviders = uint32(pb.KademliaMessageType_GET_PROVIDERS)
	KademliaMessageTypeFindNode     = uint32(pb.KademliaMessageType_FIND_NODE)
	KademliaMessageTypePing         = uint32(pb.KademliaMessageType_PING)
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

type KademliaQuery struct {
	Timestamp time.Time           `json:"timestamp"`
	Peer      peer.ID             `json:"peer"`
	QueryType KademliaMessageType `json:"query_type"`
	Duration  time.Duration       `json:"duration"`
}

type KademliaHandler struct {
	Timestamp       time.Time           `json:"timestamp"`
	HandlerType     KademliaMessageType `json:"handler_type"`
	HandlerDuration time.Duration       `json:"handler_duration"`
	WriteDuration   time.Duration       `json:"write_duration"`
}

func KademliaSerialize(in *Kademlia, stream *telemetry.Stream) error {
	dp := &pb.Kademlia{
		Timestamp:   pbutils.TimeToPB(&in.Timestamp),
		MessagesIn:  in.MessagesIn,
		MessagesOut: in.MessagesOut,
	}
	return stream.AllocAndWrite(dp.Size(), func(b []byte) error {
		_, err := dp.MarshalToSizedBuffer(b)
		return err
	})
}

func KademliaDeserialize(in []byte) (*Kademlia, error) {
	var inpb pb.Kademlia
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}
	return &Kademlia{
		Timestamp:   pbutils.TimeFromPB(inpb.GetTimestamp()),
		MessagesIn:  kademliaCountMapFromPB(inpb.GetMessagesIn()),
		MessagesOut: kademliaCountMapFromPB(inpb.GetMessagesOut()),
	}, nil
}

func KademliaQuerySerialize(in *KademliaQuery, stream *telemetry.Stream) error {
	inpb := &pb.KademliaQuery{
		Timestamp: pbutils.TimeToPB(&in.Timestamp),
		Peer:      in.Peer.Pretty(),
		QueryType: pb.KademliaMessageType(in.QueryType),
		Duration:  pbutils.DurationToPB(&in.Duration),
	}
	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func KademliaQueryDeserialize(in []byte) (*KademliaQuery, error) {
	var inpb pb.KademliaQuery
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	p, err := peer.Decode(inpb.GetPeer())
	if err != nil {
		return nil, err
	}

	return &KademliaQuery{
		Timestamp: pbutils.TimeFromPB(inpb.GetTimestamp()),
		Peer:      p,
		QueryType: KademliaMessageType(inpb.GetQueryType()),
		Duration:  pbutils.DurationFromPB(inpb.GetDuration()),
	}, nil
}

func KademliaHandlerSerialize(in *KademliaHandler, stream *telemetry.Stream) error {
	inpb := &pb.KademliaHandler{
		Timestamp:       pbutils.TimeToPB(&in.Timestamp),
		HandlerType:     pb.KademliaMessageType(in.HandlerType),
		HandlerDuration: pbutils.DurationToPB(&in.HandlerDuration),
		WriteDuration:   pbutils.DurationToPB(&in.WriteDuration),
	}
	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func KademliaHandlerDeserialize(in []byte) (*KademliaHandler, error) {
	var inpb pb.KademliaHandler
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	return &KademliaHandler{
		Timestamp:       pbutils.TimeFromPB(inpb.GetTimestamp()),
		HandlerType:     KademliaMessageType(inpb.GetHandlerType()),
		HandlerDuration: pbutils.DurationFromPB(inpb.GetHandlerDuration()),
		WriteDuration:   pbutils.DurationFromPB(inpb.GetWriteDuration()),
	}, nil
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
