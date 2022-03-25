package telemetry

import (
	"time"

	"git.d464.sh/adc/telemetry/telemetry/snapshot"
)

const COLLECTOR_RT_INTERVAL = time.Second * 30

func (t *TelemetryService) collectorRT() {
	for {
		snapshot := snapshot.NewRoutingTableSnapshotFromNode(t.n)
		t.pushSnapshot(snapshot)
		time.Sleep(COLLECTOR_RT_INTERVAL)
	}
}
