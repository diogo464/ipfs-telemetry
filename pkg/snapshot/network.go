package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/pbutils"
	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/protocol"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Network struct {
	Timestamp   time.Time                     `json:"timestamp"`
	Overall     metrics.Stats                 `json:"overall"`
	PerProtocol map[protocol.ID]metrics.Stats `json:"perprotocol"`
	NumConns    uint32                        `json:"numconns"`
	LowWater    uint32                        `json:"lowwater"`
	HighWater   uint32                        `json:"highwater"`
}

func NetworkFromPB(in *pb.Network) (*Network, error) {
	// TODO: Fix this
	return &Network{
		Timestamp: in.GetTimestamp().AsTime(),
		NumConns:  in.GetNumConns(),
		LowWater:  in.GetLowWater(),
		HighWater: in.GetHighWater(),
	}, nil
}

func (n *Network) ToPB() *pb.Network {
	byprotocol := make(map[string]*pb.Network_Stats)
	for k, v := range n.PerProtocol {
		byprotocol[string(k)] = pbutils.MetricsStatsToPB(&v)
	}

	return &pb.Network{
		Timestamp:       timestamppb.New(n.Timestamp),
		StatsOverall:    pbutils.MetricsStatsToPB(&n.Overall),
		StatsByProtocol: byprotocol,
		NumConns:        n.NumConns,
		LowWater:        n.LowWater,
		HighWater:       n.HighWater,
	}
}

func NetworkArrayToPB(in []*Network) []*pb.Network {
	out := make([]*pb.Network, 0, len(in))
	for _, p := range in {
		out = append(out, p.ToPB())
	}
	return out
}
