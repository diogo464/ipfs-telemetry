package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*Bitswap)(nil)

const BitswapName = "bitswap"

type Bitswap struct {
	Timestamp          time.Time `json:"timestamp"`
	DiscoverySucceeded uint32    `json:"discovery_succeeded"`
	DiscoveryFailed    uint32    `json:"discovery_failed"`
	MessagesIn         uint64    `json:"messages_in"`
	MessagesOut        uint64    `json:"messages_out"`
}

func (*Bitswap) sealed()                   {}
func (*Bitswap) GetName() string           { return BitswapName }
func (b *Bitswap) GetTimestamp() time.Time { return b.Timestamp }
func (b *Bitswap) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 2*4 + 2*8
}
func (b *Bitswap) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Bitswap{
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
