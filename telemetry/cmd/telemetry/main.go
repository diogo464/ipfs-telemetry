package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

var FLAGS = []cli.Flag{
	FLAG_CONN_TYPE,
	FLAG_HOST,
}

var COMMANDS = []*cli.Command{
	CommandUpload,
	CommandDownload,
	CommandSession,
	CommandMetrics,
	CommandMetricsToVm,
	CommandEvent,
	CommandProperties,
	CommandDescriptors,
}

func main() {
	app := &cli.App{
		Name:     "telemetry",
		Flags:    FLAGS,
		Commands: COMMANDS,
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
