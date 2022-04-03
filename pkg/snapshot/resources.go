package snapshot

import (
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ Snapshot = (*Resources)(nil)

const RESOURCES_NAME = "resources"

type Resources struct {
	Timestamp   time.Time `json:"timestamp"`
	CpuUsage    float32   `json:"cpu_usage"`
	MemoryUsed  uint64    `json:"memory_used"`
	MemoryFree  uint64    `json:"memory_free"`
	MemoryTotal uint64    `json:"memory_total"`
	Goroutines  uint32    `json:"goroutines"`
}

func (*Resources) sealed()                   {}
func (*Resources) GetName() string           { return RESOURCES_NAME }
func (r *Resources) GetTimestamp() time.Time { return r.Timestamp }
func (r *Resources) ToPB() *pb.Snapshot {
	return &pb.Snapshot{
		Body: &pb.Snapshot_Resources{
			Resources: ResourcesToPB(r),
		},
	}
}

func ResourcesFromPB(in *pb.Resources) (*Resources, error) {
	return &Resources{
		Timestamp:   in.GetTimestamp().AsTime(),
		CpuUsage:    in.GetCpuUsage(),
		MemoryUsed:  in.GetMemoryUsed(),
		MemoryFree:  in.GetMemoryFree(),
		MemoryTotal: in.GetMemoryTotal(),
		Goroutines:  in.GetGoroutines(),
	}, nil
}

func ResourcesToPB(r *Resources) *pb.Resources {
	return &pb.Resources{
		Timestamp:   timestamppb.New(r.Timestamp),
		CpuUsage:    r.CpuUsage,
		MemoryUsed:  r.MemoryUsed,
		MemoryFree:  r.MemoryFree,
		MemoryTotal: r.MemoryTotal,
		Goroutines:  r.Goroutines,
	}
}
