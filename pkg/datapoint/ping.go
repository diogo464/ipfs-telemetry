package datapoint

import (
	"time"

	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	"github.com/libp2p/go-libp2p-core/peer"
)

const PingName = "ping"

type Ping struct {
	Timestamp   time.Time       `json:"timestamp"`
	Source      peer.AddrInfo   `json:"source"`
	Destination peer.AddrInfo   `json:"destination"`
	Durations   []time.Duration `json:"durations"`
}

func PingSerialize(in *Ping, stream *telemetry.Stream) error {
	source := pbutils.AddrInfoToPB(&in.Source)
	destination := pbutils.AddrInfoToPB(&in.Destination)

	inpb := &pb.Ping{
		Timestamp:   pbutils.TimeToPB(&in.Timestamp),
		Source:      source,
		Destination: destination,
		Durations:   pbutils.DurationArrayToPB(in.Durations),
	}

	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func PingDeserialize(in []byte) (*Ping, error) {
	var inpb pb.Ping
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	source, err := pbutils.AddrInfoFromPB(inpb.GetSource())
	if err != nil {
		return nil, err
	}
	dest, err := pbutils.AddrInfoFromPB(inpb.GetDestination())
	if err != nil {
		return nil, err
	}
	durations := make([]time.Duration, 0, len(inpb.GetDurations()))
	for _, dur := range inpb.Durations {
		durations = append(durations, pbutils.DurationFromPB(dur))
	}
	return &Ping{
		Timestamp:   pbutils.TimeFromPB(inpb.GetTimestamp()),
		Source:      source,
		Destination: dest,
		Durations:   durations,
	}, nil
}
