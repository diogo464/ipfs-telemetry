package collector

import (
	"context"
	"os"
	"runtime"

	"github.com/diogo464/telemetry/pkg/telemetry/datapoint"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

const INVALID_CPU_USAGE float64 = -1.0
const INVALID_MEM_VALUE uint64 = 0

var _ Collector = (*resourcesCollector)(nil)

type resourcesCollector struct {
	numCPU int
	proc   *process.Process
	stats  *runtime.MemStats
}

func NewResourcesCollector() (Collector, error) {
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

// Close implements Collector
func (*resourcesCollector) Close() {
}

// Collect implements Collector
func (c *resourcesCollector) Collect(ctx context.Context, sink datapoint.Sink) {
	var cpuProcess float64 = INVALID_CPU_USAGE
	var cpuSystem float64 = INVALID_CPU_USAGE

	cpuProcess, err := c.proc.PercentWithContext(ctx, 0)
	if err != nil {
		cpuProcess = INVALID_CPU_USAGE
	} else {
		cpuProcess /= float64(c.numCPU)
	}

	cpuSystemUsage, err := cpu.Percent(0, false)
	if err == nil {
		cpuSystem = cpuSystemUsage[0]
	}

	var mem_free uint64 = INVALID_MEM_VALUE
	var mem_total uint64 = INVALID_MEM_VALUE
	memory, err := mem.VirtualMemory()
	if err == nil {
		mem_free = memory.Free
		mem_total = memory.Total
	}

	runtime.ReadMemStats(c.stats)
	sink.Push(&datapoint.Resources{
		Timestamp:   datapoint.NewTimestamp(),
		CpuProcess:  float32(cpuProcess),
		CpuSystem:   float32(cpuSystem),
		MemoryUsed:  c.stats.HeapAlloc,
		MemoryFree:  mem_free,
		MemoryTotal: mem_total,
		Goroutines:  uint32(runtime.NumGoroutine()),
	})
}
