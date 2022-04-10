package crawler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PeersCurrentCrawl = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_peers_current_crawl",
		Help: "Number unique peers found in this crawl",
	})
	PeersLastCrawl = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_peers_last_crawl",
		Help: "Number of unique peers found in the last crawl",
	})
	PeersTelemetryCurrentCrawl = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_peers_telemetry_current_crawl",
		Help: "Number of unique peers found in this crawl supporting the telemetry protocol",
	})
	PeersTelemetryLastCrawl = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_peers_telemetry_last_crawl",
		Help: "Number of unique peers found in the last crawl supporting the telemetry protocol",
	})
	ErrorsCurrentCrawl = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_errors_current_crawl",
		Help: "Number of errors in this crawl",
	})
	ErrorsLastCrawl = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_errors_last_crawl",
		Help: "Number of errros in the last crawl",
	})
	CompletedCrawls = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_completed_crawls",
		Help: "Number of completed crawls",
	})
)
