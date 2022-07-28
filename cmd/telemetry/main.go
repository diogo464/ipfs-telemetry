package main

import (
	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	telemetry_cli "github.com/diogo464/telemetry/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	for name, decoder := range datapoint.Decoders {
		telemetry_cli.RegisterStreamDecoder(name, decoder)
	}
	app := &cli.App{
		Name:     "telemetry",
		Flags:    telemetry_cli.FLAGS,
		Commands: telemetry_cli.COMMANDS,
	}
	app.RunAndExitOnError()
}
