package monitor

import (
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Exporter interface {
	Export(peer.ID, []snapshot.Snapshot)
}

type fnExporter struct {
	fn func(peer.ID, []snapshot.Snapshot)
}

func (e *fnExporter) Export(p peer.ID, s []snapshot.Snapshot) {
	e.fn(p, s)
}

func NewExporterFn(fn func(peer.ID, []snapshot.Snapshot)) Exporter {
	return &fnExporter{fn}
}
