package main

import (
	"net"
	"os"
	"strings"

	"git.d464.sh/adc/telemetry/pkg/orchestrator"
	"git.d464.sh/adc/telemetry/pkg/probe"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var _ orchestrator.Exporter = (*influxExporter)(nil)

type influxExporter struct {
	writer api.WriteAPI
}

// Export implements orchestrator.Exporter
func (e *influxExporter) Export(n string, r *probe.ProbeResult) {
	if r.Error != nil {
		return
	}

	p := influxdb2.NewPointWithMeasurement("probe").
		AddTag("peer", r.Peer.String()).
		AddTag("probe", n).
		AddField("duration", r.RequestDuration.Nanoseconds()).
		SetTime(r.RequestStart)
	e.writer.WritePoint(p)
}

func main() {
	app := &cli.App{
		Name:     "orchestrator",
		Commands: []*cli.Command{},
		Flags: []cli.Flag{
			FLAG_PROBES,
			FLAG_NUM_CIDS,
			FLAG_INFLUXDB_ADDRESS,
			FLAG_INFLUXDB_TOKEN,
			FLAG_INFLUXDB_ORG,
			FLAG_INFLUXDB_BUCKET,
		},
		Action: mainAction,
	}

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	numCids := c.Int(FLAG_NUM_CIDS.Name)
	logrus.Debug("num cids = ", numCids)

	probes := make([]net.Addr, 0)
	for _, paddr := range strings.Split(c.String(FLAG_PROBES.Name), ",") {
		addr, err := net.ResolveTCPAddr("tcp", paddr)
		if err != nil {
			return err
		}
		probes = append(probes, addr)
	}

	client := influxdb2.NewClient(c.String(FLAG_INFLUXDB_ADDRESS.Name), c.String(FLAG_INFLUXDB_TOKEN.Name))
	defer client.Close()
	writeAPI := client.WriteAPI(c.String(FLAG_INFLUXDB_ORG.Name), c.String(FLAG_INFLUXDB_BUCKET.Name))
	exporter := &influxExporter{writeAPI}
	defer writeAPI.Flush()

	server, err := orchestrator.NewOrchestratorServer(
		c.Context,
		orchestrator.WithExporter(exporter),
		orchestrator.WithProbes(probes),
		orchestrator.WithNumCids(numCids),
	)
	if err != nil {
		return err
	}

	return server.Run(c.Context)
}
