package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TELEMETRY_REQUESTS_TOTAL = promauto.NewCounter(prometheus.CounterOpts{
		Name:        "monitor_telemetry_requests_total",
		Help:        "The total number of telemetry requests made, successful or not",
		ConstLabels: map[string]string{},
	})
)
