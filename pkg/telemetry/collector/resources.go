package collector

import (
	"context"
	"runtime"

	"git.d464.sh/adc/telemetry/pkg/telemetry/snapshot"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const INVALID_CPU_USAGE float64 = -1.0
const INVALID_MEM_VALUE uint64 = 0

var _ Collector = (*resourcesCollector)(nil)

type resourcesCollector struct {
	stats *runtime.MemStats
}

func NewResourcesCollector() Collector {
	return &resourcesCollector{
		stats: new(runtime.MemStats),
	}
}

// Close implements Collector
func (*resourcesCollector) Close() {
}

// Collect implements Collector
func (c *resourcesCollector) Collect(ctx context.Context, sink snapshot.Sink) {
	var usage float64 = INVALID_CPU_USAGE
	usagearr, err := cpu.Percent(0, false)
	if err == nil {
		usage = usagearr[0]
	}

	var mem_free uint64 = INVALID_MEM_VALUE
	var mem_total uint64 = INVALID_MEM_VALUE
	memory, err := mem.VirtualMemory()
	if err == nil {
		mem_free = memory.Free
		mem_total = memory.Total
	}

	runtime.ReadMemStats(c.stats)
	sink.Push(&snapshot.Resources{
		Timestamp:   snapshot.NewTimestamp(),
		CpuUsage:    float32(usage),
		MemoryUsed:  c.stats.HeapAlloc,
		MemoryFree:  mem_free,
		MemoryTotal: mem_total,
		Goroutines:  uint32(runtime.NumGoroutine()),
	})
}
