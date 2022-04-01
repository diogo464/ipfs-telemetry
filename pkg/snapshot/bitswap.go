package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Bitswap struct {
	Timestamp          time.Time `json:"timestamp"`
	DiscoverySucceeded uint32    `json:"discovery_succeeded"`
	DiscoveryFailed    uint32    `json:"discovery_failed"`
}

func BitswapFromPB(in *pb.Bitswap) (*Bitswap, error) {
	return &Bitswap{
		Timestamp:          in.Timestamp.AsTime(),
		DiscoverySucceeded: in.GetDiscoverySucceeded(),
		DiscoveryFailed:    in.GetDiscoveryFailed(),
	}, nil
}

func (bs *Bitswap) ToPB() *pb.Bitswap {
	return &pb.Bitswap{
		Timestamp:          timestamppb.New(bs.Timestamp),
		DiscoverySucceeded: bs.DiscoverySucceeded,
		DiscoveryFailed:    bs.DiscoveryFailed,
	}
}