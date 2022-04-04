package telemetry

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"git.d464.sh/adc/telemetry/pkg/telemetry/window"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func metricsTask(w window.Window) {
	snapshotMemory := make(map[string]prometheus.Gauge)
	snapshotCounts := make(map[string]prometheus.Gauge)

	go func() {
		for {
			for _, v := range snapshotMemory {
				v.Set(0)
			}
			for _, v := range snapshotCounts {
				v.Set(0)
			}

			for _, snapshotpb := range w.Since(0) {
				s, err := snapshot.FromPB(snapshotpb)
				if err != nil {
					panic(err)
				}
				n := s.GetName()

				if _, ok := snapshotMemory[n]; !ok {
					snapshotMemory[n] = promauto.NewGauge(prometheus.GaugeOpts{
						Name:        "telemetry_snapshot_memory",
						Help:        "Memory used by snapshots",
						ConstLabels: map[string]string{"kind": n},
					})
					snapshotCounts[n] = promauto.NewGauge(prometheus.GaugeOpts{
						Name:        "telemetry_snapshot_count",
						Help:        "Count of snapshots",
						ConstLabels: map[string]string{"kind": n},
					})
				}

				snapshotMemory[n].Add(float64(snapshot.SnapshotSize(s)))
				snapshotCounts[n].Inc()
			}
			time.Sleep(time.Second * 4)
		}
	}()
}
