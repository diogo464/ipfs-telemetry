package telemetry

import mpb "go.opentelemetry.io/proto/otlp/metrics/v1"

type MetricDescriptor struct {
	Scope       string `json:"scope"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Metrics struct {
	OTLP []*mpb.ResourceMetrics `json:"otlp"`
}
