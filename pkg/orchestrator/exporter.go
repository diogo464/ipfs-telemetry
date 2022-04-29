package orchestrator

import "github.com/diogo464/telemetry/pkg/probe"

var _ (Exporter) = (*NullExporter)(nil)

type Exporter interface {
	Export(probeName string, result *probe.ProbeResult)
}

type NullExporter struct{}

// Export implements Exporter
func (*NullExporter) Export(probeName string, result *probe.ProbeResult) {
}
