package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*KademliaQuery)(nil)

const KADEMLIA_QUERY_NAME = "kademliaquery"

type KademliaQuery struct {
	Timestamp time.Time     `json:"timestamp"`
	Peer      peer.ID       `json:"peer"`
	Duration  time.Duration `json:"duration"`
}

func (*KademliaQuery) sealed()                   {}
func (*KademliaQuery) GetName() string           { return KADEMLIA_QUERY_NAME }
func (p *KademliaQuery) GetTimestamp() time.Time { return p.Timestamp }
func (p *KademliaQuery) ToPB() *pb.Snapshot {
	return &pb.Snapshot{
		Body: &pb.Snapshot_KademliaQuery{
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
		Timestamp: in.GetTimestamp().AsTime(),
		Peer:      p,
		Duration:  in.GetDuration().AsDuration(),
	}, nil
}

func KademliaQueryToPB(p *KademliaQuery) *pb.KademliaQuery {
	return &pb.KademliaQuery{
		Timestamp: timestamppb.New(p.Timestamp),
		Peer:      p.Peer.Pretty(),
		Duration:  durationpb.New(p.Duration),
	}
}
