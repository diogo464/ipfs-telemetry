package telemetry

import mpb "go.opentelemetry.io/proto/otlp/metrics/v1"

type MetricDescriptor struct {
	Scope       string
	Name        string
	Description string
}

type Metrics struct {
	OTLP []*mpb.ResourceMetrics
}
