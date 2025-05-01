package main

import (
	"os"

	"github.com/diogo464/ipfs-telemetry/backend"
	"github.com/diogo464/ipfs-telemetry/backend/crawler"
	"github.com/diogo464/ipfs-telemetry/backend/monitor"
	"github.com/diogo464/ipfs-telemetry/backend/pg_crawler_exporter"
	"github.com/diogo464/ipfs-telemetry/backend/vm_otlp_exporter"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			backend.Flag_PrometheusAddress,
			backend.Flag_VmUrl,
			backend.Flag_NatsUrl,
			backend.Flag_PostgresUrl,
		},
		Commands: []*cli.Command{
			crawler.Command,
			monitor.Command,
			vm_otlp_exporter.Command,
			pg_crawler_exporter.Command,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
