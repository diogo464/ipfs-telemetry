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
	CommandEvent,
	CommandProperties,
	CommandWalk,
	CommandDescriptors,
}

func main() {
	app := &cli.App{
		Name:     "telemetry",
		Flags:    FLAGS,
		Commands: COMMANDS,
	}
	app.Run(os.Args)
}
