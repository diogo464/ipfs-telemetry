package main

import (
	telemetry_cli "github.com/diogo464/telemetry/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:     "telemetry",
		Flags:    telemetry_cli.FLAGS,
		Commands: telemetry_cli.COMMANDS,
	}
	app.RunAndExitOnError()
}
