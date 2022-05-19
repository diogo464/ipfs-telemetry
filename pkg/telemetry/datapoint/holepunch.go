package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Datapoint = (*HolePunch)(nil)

const HolePunchName = "holepunch"

type HolePunch struct {
	Timestamp       time.Time `json:"timestamp"`
	IncomingSuccess uint32    `json:"incoming_success"`
	IncomingFailure uint32    `json:"incoming_failure"`
	OutgoingSuccess uint32    `json:"outgoing_success"`
	OutgoingFailure uint32    `json:"outgoing_failure"`
}

func (*HolePunch) sealed()                   {}
func (*HolePunch) GetName() string           { return HolePunchName }
func (b *HolePunch) GetTimestamp() time.Time { return b.Timestamp }
func (b *HolePunch) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 4*4
}
func (c *HolePunch) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Holepunch{
			Holepunch: HolePunchToPB(c),
		},
	}
}

func HolePunchFromPB(in *pb.HolePunch) (*HolePunch, error) {
	return &HolePunch{
		Timestamp:       in.GetTimestamp().AsTime(),
		IncomingSuccess: in.GetIncomingSuccess(),
		IncomingFailure: in.GetIncomingFailure(),
		OutgoingSuccess: in.GetOutgoingSucess(),
		OutgoingFailure: in.GetOutgoingFailure(),
	}, nil
}

func HolePunchToPB(c *HolePunch) *pb.HolePunch {
	return &pb.HolePunch{
		Timestamp:       timestamppb.New(c.Timestamp),
		IncomingSuccess: c.IncomingSuccess,
		IncomingFailure: c.IncomingFailure,
		OutgoingSucess:  c.OutgoingSuccess,
		OutgoingFailure: c.OutgoingFailure,
	}
}
