package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/durationpb"
)

func addrInfoToPB(info *peer.AddrInfo) *pb.AddrInfo {
	a := make([]string, 0, len(info.Addrs))
	for _, addr := range info.Addrs {
		a = append(a, addr.String())
	}
	return &pb.AddrInfo{
		Id:    info.ID.Pretty(),
		Addrs: a,
	}
}

func durationsToPbDurations(in []time.Duration) []*durationpb.Duration {
	out := make([]*durationpb.Duration, 0, len(in))
	for _, dur := range in {
		out = append(out, durationpb.New(dur))
	}
	return out
}
