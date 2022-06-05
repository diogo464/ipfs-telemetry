package pbutils

import (
	"io"
	"time"

	pbc "github.com/diogo464/telemetry/pkg/proto/common"
	pbs "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/rle"
	"github.com/gogo/protobuf/types"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/protobuf/proto"
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

func TimeFromPB(in *types.Timestamp) time.Time {
	return time.Unix(in.Seconds, int64(in.Nanos))
}

func TimeToPB(in *time.Time) *types.Timestamp {
	return &types.Timestamp{
		Seconds: in.Unix(),
		Nanos:   int32(in.Nanosecond()),
	}
}

func DurationFromPB(in *types.Duration) time.Duration {
	return time.Duration(in.Seconds + int64(in.Nanos))
}

func DurationToPB(in *time.Duration) *types.Duration {
	return &types.Duration{
		Seconds: in.Nanoseconds() / 1_000_000_000,
		Nanos:   int32(in.Nanoseconds() % 1_000_000_000),
	}
}

func DurationArrayToPB(in []time.Duration) []*types.Duration {
	out := make([]*types.Duration, 0, len(in))
	for _, dur := range in {
		out = append(out, DurationToPB(&dur))
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

func MultiAddrsToPB(in []multiaddr.Multiaddr) []string {
	addrs := make([]string, 0, len(in))
	for _, a := range in {
		addrs = append(addrs, a.String())
	}
	return addrs
}

func ReadRle(r io.Reader, v proto.Message) error {
	data, err := rle.Read(r)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, v)
}
