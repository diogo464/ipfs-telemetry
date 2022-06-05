package datapoint

import (
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry/pbutils"
)

var _ Datapoint = (*Resources)(nil)

const ResourceName = "resources"

type Resources struct {
	Timestamp   time.Time `json:"timestamp"`
	CpuProcess  float32   `json:"cpu_process"`
	CpuSystem   float32   `json:"cpu_system"`
	MemoryUsed  uint64    `json:"memory_used"`
	MemoryFree  uint64    `json:"memory_free"`
	MemoryTotal uint64    `json:"memory_total"`
	Goroutines  uint32    `json:"goroutines"`
}

func (*Resources) sealed()                   {}
func (*Resources) GetName() string           { return ResourceName }
func (r *Resources) GetTimestamp() time.Time { return r.Timestamp }
func (r *Resources) GetSizeEstimate() uint32 {
	return estimateTimestampSize + 2*4 + 3*8 + 4
}
func (r *Resources) ToPB() *pb.Datapoint {
	return &pb.Datapoint{
		Body: &pb.Datapoint_Resources{
			Resources: ResourcesToPB(r),
		},
	}
}

func ResourcesFromPB(in *pb.Resources) (*Resources, error) {
	return &Resources{
		Timestamp:   pbutils.TimeFromPB(in.GetTimestamp()),
		CpuProcess:  in.GetCpuProcess(),
		CpuSystem:   in.GetCpuSystem(),
		MemoryUsed:  in.GetMemoryUsed(),
		MemoryFree:  in.GetMemoryFree(),
		MemoryTotal: in.GetMemoryTotal(),
		Goroutines:  in.GetGoroutines(),
	}, nil
}

func ResourcesToPB(r *Resources) *pb.Resources {
	return &pb.Resources{
		Timestamp:   pbutils.TimeToPB(&r.Timestamp),
		CpuProcess:  r.CpuProcess,
		CpuSystem:   r.CpuSystem,
		MemoryUsed:  r.MemoryUsed,
		MemoryFree:  r.MemoryFree,
		MemoryTotal: r.MemoryTotal,
		Goroutines:  r.Goroutines,
	}
}
