package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*Bitswap)(nil)

const BITSWAP_NAME = "bitswap"

type Bitswap struct {
	Timestamp          time.Time `json:"timestamp"`
	DiscoverySucceeded uint32    `json:"discovery_succeeded"`
	DiscoveryFailed    uint32    `json:"discovery_failed"`
	MessagesIn         uint32    `json:"messages_in"`
	MessagesOut        uint32    `json:"messages_out"`
}

func (*Bitswap) sealed()                   {}
func (*Bitswap) GetName() string           { return BITSWAP_NAME }
func (b *Bitswap) GetTimestamp() time.Time { return b.Timestamp }
func (b *Bitswap) ToPB() *pb.Snapshot {
	return &pb.Snapshot{
		Body: &pb.Snapshot_Bitswap{
			Bitswap: BitswapToPB(b),
		},
	}
}

func BitswapFromPB(in *pb.Bitswap) (*Bitswap, error) {
	return &Bitswap{
		Timestamp:          in.Timestamp.AsTime(),
		DiscoverySucceeded: in.GetDiscoverySucceeded(),
		DiscoveryFailed:    in.GetDiscoveryFailed(),
		MessagesIn:         in.GetMessagesIn(),
		MessagesOut:        in.GetMessagesOut(),
	}, nil
}

func BitswapToPB(bs *Bitswap) *pb.Bitswap {
	return &pb.Bitswap{
		Timestamp:          timestamppb.New(bs.Timestamp),
		DiscoverySucceeded: bs.DiscoverySucceeded,
		DiscoveryFailed:    bs.DiscoveryFailed,
		MessagesIn:         bs.MessagesIn,
		MessagesOut:        bs.MessagesOut,
	}
}
