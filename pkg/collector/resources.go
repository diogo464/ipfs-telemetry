package collector

import (
	"context"
	"runtime"
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const INVALID_CPU_USAGE float64 = -1.0
const INVALID_MEM_VALUE uint64 = 0

type ResourcesOptions struct {
	Interval time.Duration
}

type resourcesCollector struct {
	ctx  context.Context
	opts ResourcesOptions
	sink snapshot.Sink
}

func RunResourcesCollector(ctx context.Context, sink snapshot.Sink, opts ResourcesOptions) {
	c := &resourcesCollector{ctx, opts, sink}
	c.Run()
}

func (c *resourcesCollector) Run() {
	ticker := time.NewTicker(c.opts.Interval)
	stats := new(runtime.MemStats)

LOOP:
	for {
		select {
		case <-ticker.C:
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

			runtime.ReadMemStats(stats)
			c.sink.PushResources(&snapshot.Resources{
				Timestamp:   snapshot.NewTimestamp(),
				CpuUsage:    float32(usage),
				MemoryUsed:  stats.HeapAlloc,
				MemoryFree:  mem_free,
				MemoryTotal: mem_total,
				Goroutines:  uint32(runtime.NumGoroutine()),
			})
		case <-c.ctx.Done():
			break LOOP
		}
	}
}
