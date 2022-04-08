package telemetry

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/window"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func metricsTask(w window.Window) {
	snapshotMemory := make(map[string]prometheus.Gauge)
	snapshotCounts := make(map[string]prometheus.Gauge)
	stats := new(window.Stats)

	for {
		w.Stats(stats)

		for n, mem := range stats.MemoryCounts {
			if _, ok := snapshotMemory[n]; !ok {
				snapshotMemory[n] = promauto.NewGauge(prometheus.GaugeOpts{
					Name:        "telemetry_snapshot_memory",
					Help:        "Memory used by snapshots",
					ConstLabels: map[string]string{"kind": n},
				})
			}
			snapshotMemory[n].Set(float64(mem))
		}
		for n, count := range stats.SnapshotCounts {
			if _, ok := snapshotCounts[n]; !ok {
				snapshotCounts[n] = promauto.NewGauge(prometheus.GaugeOpts{
					Name:        "telemetry_snapshot_count",
					Help:        "Count of snapshots",
					ConstLabels: map[string]string{"kind": n},
				})
			}
			snapshotCounts[n].Set(float64(count))
		}

		time.Sleep(time.Second * 4)
	}
}
