package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
)

var _ Datapoint = (*HolePunch)(nil)

const HolePunchName = "holepunch"

type HolePunch struct {
	Timestamp time.Time `json:"timestamp"`
	Success   uint32    `json:"success"`
	Failure   uint32    `json:"failure"`
}

func (*HolePunch) sealed()                   {}
func (*HolePunch) GetName() string           { return HolePunchName }
func (b *HolePunch) GetTimestamp() time.Time { return b.Timestamp }
func (b *HolePunch) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 2*4
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
		Timestamp: pbutils.TimeFromPB(in.GetTimestamp()),
		Success:   in.GetSuccess(),
		Failure:   in.GetFailure(),
	}, nil
}

func HolePunchToPB(c *HolePunch) *pb.HolePunch {
	return &pb.HolePunch{
		Timestamp: pbutils.TimeToPB(&c.Timestamp),
		Success:   c.Success,
		Failure:   c.Failure,
	}
}
