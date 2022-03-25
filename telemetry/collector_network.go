package telemetry

import (
	"time"

	"git.d464.sh/adc/telemetry/telemetry/snapshot"
)

func (t *TelemetryService) collectorNetwork() {
	for {
		snapshot := snapshot.NewNetworkSnapshotFromNode(t.n)
		t.pushSnapshot(snapshot)
		time.Sleep(time.Second * 10)
	}
}
