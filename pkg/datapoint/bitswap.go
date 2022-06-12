package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
)

const BitswapName = "bitswap"

type Bitswap struct {
	Timestamp          time.Time `json:"timestamp"`
	DiscoverySucceeded uint32    `json:"discovery_succeeded"`
	DiscoveryFailed    uint32    `json:"discovery_failed"`
	MessagesIn         uint64    `json:"messages_in"`
	MessagesOut        uint64    `json:"messages_out"`
}

func BitswapSerialize(in *Bitswap, stream *telemetry.Stream) error {
	dp := &pb.Bitswap{
		Timestamp:          pbutils.TimeToPB(&in.Timestamp),
		DiscoverySucceeded: in.DiscoverySucceeded,
		DiscoveryFailed:    in.DiscoveryFailed,
		MessagesIn:         in.MessagesIn,
		MessagesOut:        in.MessagesOut,
	}
	return stream.AllocAndWrite(dp.Size(), func(buf []byte) error {
		_, err := dp.MarshalToSizedBuffer(buf)
		return err
	})
}

func BitswapDeserialize(in []byte) (*Bitswap, error) {
	var inpb pb.Bitswap
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}
	return &Bitswap{
		Timestamp:          pbutils.TimeFromPB(inpb.GetTimestamp()),
		DiscoverySucceeded: inpb.GetDiscoverySucceeded(),
		DiscoveryFailed:    inpb.GetDiscoveryFailed(),
		MessagesIn:         inpb.GetMessagesIn(),
		MessagesOut:        inpb.GetMessagesOut(),
	}, nil
}
