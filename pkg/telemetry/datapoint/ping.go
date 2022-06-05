package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
	"github.com/libp2p/go-libp2p-core/peer"
)

var _ Datapoint = (*Ping)(nil)

const PingName = "ping"

type Ping struct {
	Timestamp   time.Time       `json:"timestamp"`
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

func (*Ping) sealed()                   {}
func (*Ping) GetName() string           { return PingName }
func (p *Ping) GetTimestamp() time.Time { return p.Timestamp }
func (p *Ping) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 2*estimatePeerAddrInfoSize + uint32(len(p.Durations))*estimateDurationSize
}
func (p *Ping) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Ping{
			Ping: PingToPB(p),
		},
	}
}

func PingFromPB(in *pb.Ping) (*Ping, error) {
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
		durations = append(durations, pbutils.DurationFromPB(dur))
	}
	return &Ping{
		Timestamp:   pbutils.TimeFromPB(in.GetTimestamp()),
		Source:      source,
		Destination: dest,
		Durations:   durations,
	}, nil
}

func PingToPB(p *Ping) *pb.Ping {
	source := pbutils.AddrInfoToPB(&p.Source)
	destination := pbutils.AddrInfoToPB(&p.Destination)

	return &pb.Ping{
		Timestamp:   pbutils.TimeToPB(&p.Timestamp),
		Source:      source,
		Destination: destination,
		Durations:   pbutils.DurationArrayToPB(p.Durations),
	}
}

func PingArrayToPB(in []*Ping) []*pb.Ping {
	out := make([]*pb.Ping, 0, len(in))
	for _, p := range in {
		out = append(out, PingToPB(p))
	}
	return out
}
