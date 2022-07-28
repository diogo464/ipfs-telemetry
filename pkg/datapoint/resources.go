package datapoint

import (
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/pbutils"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/datapoint"
	"github.com/diogo464/telemetry"
)

// TODO(d464): rename to ResourcesName
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

func ResourcesSerialize(in *Resources, stream *telemetry.Stream) error {
	inpb := &pb.Resources{
		Timestamp:   pbutils.TimeToPB(&in.Timestamp),
		CpuProcess:  in.CpuProcess,
		CpuSystem:   in.CpuSystem,
		MemoryUsed:  in.MemoryUsed,
		MemoryFree:  in.MemoryFree,
		MemoryTotal: in.MemoryTotal,
		Goroutines:  in.Goroutines,
	}
	return stream.AllocAndWrite(inpb.Size(), func(b []byte) error {
		_, err := inpb.MarshalToSizedBuffer(b)
		return err
	})
}

func ResourcesDeserialize(in []byte) (*Resources, error) {
	var inpb pb.Resources
	err := inpb.Unmarshal(in)
	if err != nil {
		return nil, err
	}

	return &Resources{
		Timestamp:   pbutils.TimeFromPB(inpb.GetTimestamp()),
		CpuProcess:  inpb.GetCpuProcess(),
		CpuSystem:   inpb.GetCpuSystem(),
		MemoryUsed:  inpb.GetMemoryUsed(),
		MemoryFree:  inpb.GetMemoryFree(),
		MemoryTotal: inpb.GetMemoryTotal(),
		Goroutines:  inpb.GetGoroutines(),
	}, nil
}
