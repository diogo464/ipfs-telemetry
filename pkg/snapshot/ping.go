package snapshot

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/pbutils"
	"git.d464.sh/adc/telemetry/pkg/telemetry/pb"
	"github.com/libp2p/go-libp2p-core/peer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Ping struct {
	Timestamp   time.Time       `json:"timestamp"`
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

func PingFromPB(in *pb.Snapshot_Ping) (*Ping, error) {
	source, err := pbutils.AddrInfoFromPB(in.Source)
	if err != nil {
		return nil, err
	}
	dest, err := pbutils.AddrInfoFromPB(in.Destination)
	if err != nil {
		return nil, err
	}
	durations := make([]time.Duration, 0, len(in.Durations))
	for _, dur := range in.Durations {
		durations = append(durations, dur.AsDuration())
	}
	return &Ping{
		Timestamp:   in.GetTimestamp().AsTime(),
		Source:      source,
		Destination: dest,
		Durations:   durations,
	}, nil
}

func (p *Ping) ToPB() *pb.Snapshot_Ping {
	source := pbutils.AddrInfoToPB(&p.Source)
	destination := pbutils.AddrInfoToPB(&p.Destination)

	return &pb.Snapshot_Ping{
		Timestamp:   timestamppb.New(p.Timestamp),
		Source:      source,
		Destination: destination,
		Durations:   pbutils.DurationArrayToPB(p.Durations),
	}
}

func PingArrayToPB(in []*Ping) []*pb.Snapshot_Ping {
	out := make([]*pb.Snapshot_Ping, 0, len(in))
	for _, p := range in {
		out = append(out, p.ToPB())
	}
	return out
}
