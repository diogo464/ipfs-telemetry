package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	UniquePeers = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_unique_peers",
		Help: "Number of unique peers seen by this crawler",
	})
	UniquePeersTelemetry = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_unique_peers_telemetry",
		Help: "Number of unique peers, supporting the telemetry protocol, seen by this crawler",
	})
	SuccessfulCrawls = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_successful_crawls",
		Help: "Number peers crawled successfully",
	})
	FailedCrawls = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_failed_crawls",
		Help: "Number of peers that failed to be crawled",
	})
	TotalCrawls = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_total_crawls",
		Help: "Total number of crawled peers, failed or not",
	})
)
