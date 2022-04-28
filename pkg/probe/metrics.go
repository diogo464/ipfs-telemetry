package probe

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OngoingProbes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "probe_ongoing",
		Help: "Number of ongoing probes",
	})
	TotalProbes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "probe_total",
		Help: "Number of unique peers found in the last crawl supporting the telemetry protocol",
	})
	SuccessfulProbes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "probe_successful",
		Help: "Number of successful probes",
	})
	FailedProbes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "probe_failed",
		Help: "Number of failed probes",
	})
	Sessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "probe_sessions",
		Help: "Number of active sessions",
	})
)
