package collectors

import (
	"context"
	"os"
	"runtime"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

const INVALID_CPU_USAGE float64 = -1.0
const INVALID_MEM_VALUE uint64 = 0

var _ telemetry.Collector = (*resourcesCollector)(nil)

type resourcesCollector struct {
	numCPU int
	proc   *process.Process
	stats  *runtime.MemStats
}

func Resources() (telemetry.Collector, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	return &resourcesCollector{
		numCPU: runtime.NumCPU(),
		proc:   proc,
		stats:  new(runtime.MemStats),
	}, nil
}

// Descriptor implements telemetry.Collector
func (*resourcesCollector) Descriptor() telemetry.CollectorDescriptor {
	return telemetry.CollectorDescriptor{
		Name: datapoint.ResourceName,
	}
}

// Open implements telemetry.Collector
func (*resourcesCollector) Open() {
}

// Close implements Collector
func (*resourcesCollector) Close() {
}

// Collect implements Collector
func (c *resourcesCollector) Collect(ctx context.Context, stream *telemetry.Stream) error {
	var cpuProcess float64
	var cpuSystem float64

	cpuProcess, err := c.proc.PercentWithContext(ctx, 0)
	if err != nil {
		cpuProcess = INVALID_CPU_USAGE
	} else {
		cpuProcess /= float64(c.numCPU)
	}

	cpuSystemUsage, err := cpu.Percent(0, false)
	if err != nil {
		cpuSystem = INVALID_CPU_USAGE
	} else {
		cpuSystem = cpuSystemUsage[0]
	}

	var mem_free uint64 = INVALID_MEM_VALUE
	var mem_total uint64 = INVALID_MEM_VALUE
	memory, err := mem.VirtualMemory()
	if err == nil {
		mem_free = memory.Available
		mem_total = memory.Total
	}

	runtime.ReadMemStats(c.stats)
	dp := &datapoint.Resources{
		Timestamp:   datapoint.NewTimestamp(),
		CpuProcess:  float32(cpuProcess),
		CpuSystem:   float32(cpuSystem),
		MemoryUsed:  c.stats.HeapAlloc,
		MemoryFree:  mem_free,
		MemoryTotal: mem_total,
		Goroutines:  uint32(runtime.NumGoroutine()),
	}
	return datapoint.ResourcesSerialize(dp, stream)
}
