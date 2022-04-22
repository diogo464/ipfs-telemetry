package telemetry

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry/window"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func metricsTask(w window.Window) {
	datapointMemory := make(map[string]prometheus.Gauge)
	datapointCounts := make(map[string]prometheus.Gauge)
	stats := new(window.Stats)

	for {
		w.Stats(stats)

		for n, mem := range stats.Memory {
			if _, ok := datapointMemory[n]; !ok {
				datapointMemory[n] = promauto.NewGauge(prometheus.GaugeOpts{
					Name:        "telemetry_datapoint_memory",
					Help:        "Memory used by datapoints",
					ConstLabels: map[string]string{"kind": n},
				})
			}
			datapointMemory[n].Set(float64(mem))
		}
		for n, count := range stats.Count {
			if _, ok := datapointCounts[n]; !ok {
				datapointCounts[n] = promauto.NewGauge(prometheus.GaugeOpts{
					Name:        "telemetry_datapoint_count",
					Help:        "Count of datapoints",
					ConstLabels: map[string]string{"kind": n},
				})
			}
			datapointCounts[n].Set(float64(count))
		}

		time.Sleep(time.Second * 4)
	}
}
