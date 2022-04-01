package pbutils

import (
	"time"

	pbc "git.d464.sh/adc/telemetry/pkg/proto/common"
	pbs "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/types/known/durationpb"
)

func AddrInfoFromPB(in *pbc.AddrInfo) (peer.AddrInfo, error) {
	i, err := peer.Decode(in.Id)
	if err != nil {
		return peer.AddrInfo{}, err
	}

	a := make([]multiaddr.Multiaddr, 0, len(in.Addrs))
	for _, addr := range in.Addrs {
		x, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return peer.AddrInfo{}, err
		}
		a = append(a, x)
	}

	return peer.AddrInfo{ID: i, Addrs: a}, nil
}

func AddrInfoToPB(in *peer.AddrInfo) *pbc.AddrInfo {
	a := make([]string, 0, len(in.Addrs))
	for _, addr := range in.Addrs {
		a = append(a, addr.String())
	}
	return &pbc.AddrInfo{
		Id:    in.ID.Pretty(),
		Addrs: a,
	}
}

func DurationArrayToPB(in []time.Duration) []*durationpb.Duration {
	out := make([]*durationpb.Duration, 0, len(in))
	for _, dur := range in {
		out = append(out, durationpb.New(dur))
	}
	return out
}

func MetricsStatsToPB(in *metrics.Stats) *pbs.Network_Stats {
	return &pbs.Network_Stats{
		TotalIn:  uint64(in.TotalIn),
		TotalOut: uint64(in.TotalOut),
		RateIn:   uint64(in.RateIn),
		RateOut:  uint64(in.RateOut),
	}
}

func MetricsStatsFromPB(in *pbs.Network_Stats) metrics.Stats {
	return metrics.Stats{
		TotalIn:  int64(in.GetTotalIn()),
		TotalOut: int64(in.GetTotalOut()),
		RateIn:   float64(in.GetRateIn()),
		RateOut:  float64(in.GetRateOut()),
	}
}
