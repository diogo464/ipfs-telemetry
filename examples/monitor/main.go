package main

import (
	"fmt"
	"os"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/monitor"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli/v2"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

var _ (monitor.Exporter) = (*printExporter)(nil)

type printExporter struct{}

// Bandwidth implements monitor.Exporter
func (*printExporter) Bandwidth(p peer.ID, b telemetry.Bandwidth) {
	fmt.Println("Exporting bandwidth for peer", p)
	fmt.Printf("\tIn: %d\n", b.UploadRate)
	fmt.Printf("\tOut: %d\n", b.DownloadRate)
}

// Captures implements monitor.Exporter
func (*printExporter) Captures(p peer.ID, sess telemetry.Session, d telemetry.CaptureDescriptor, c []monitor.Capture) {
	fmt.Println("Exporting captures")
	fmt.Println("Descriptor:")
	fmt.Printf("\tName: %s\n", d.Name)
	fmt.Printf("\tDescription: %s\n", d.Description)
	fmt.Printf("\tCount: %d\n", len(c))
}

// Events implements monitor.Exporter
func (*printExporter) Events(p peer.ID, sess telemetry.Session, d telemetry.EventDescriptor, e []monitor.Event) {
	fmt.Println("Exporting events")
	fmt.Println("Descriptor:")
	fmt.Printf("\tName: %s\n", d.Name)
	fmt.Printf("\tDescription: %s\n", d.Description)
	fmt.Printf("\tCount: %d\n", len(e))
}

// Metrics implements monitor.Exporter
func (*printExporter) Metrics(peer.ID, telemetry.Session, []*mpb.ResourceMetrics) {
}

// Session implements monitor.Exporter
func (*printExporter) Session(p peer.ID, sess telemetry.Session, props []telemetry.CProperty) {
	fmt.Println("Exporting session for peer", p)
	fmt.Println("Properties:")
	for _, p := range props {
		fmt.Printf("\t%s: %s\n", p.Name, p.Value)
	}
}

var app = &cli.App{
	Name:   "monitor",
	Action: actionMain,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func actionMain(c *cli.Context) error {
	m, err := monitor.Start(c.Context,
		monitor.WithExporter(&printExporter{}),
		monitor.WithBandwidthPeriod(time.Second*10), monitor.WithCollectPeriod(time.Second*5))
	if err != nil {
		return err
	}
	addr, err := peer.AddrInfoFromString(c.Args().First())
	if err != nil {
		return err
	}
	m.DiscoverWithAddr(c.Context, *addr)
	<-c.Context.Done()
	return nil
}
