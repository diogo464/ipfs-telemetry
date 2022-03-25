package telemetry

import (
	"time"

	"d464.sh/telemetry/snapshot"
)

func (t *TelemetryService) collectorNetwork() {
	for {
		snapshot := snapshot.NewNetworkSnapshotFromNode(t.n)
		t.pushSnapshot(snapshot)
		time.Sleep(time.Second * 10)
	}
}
