package main

import (
	"os"

	"git.d464.sh/adc/telemetry/monitor/cmd/server"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "monitor",
		Description: "collect telemetry from ipfs nodes",
		Commands:    []*cli.Command{server.ServerCommand},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
