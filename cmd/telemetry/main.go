package main

import (
	"github.com/urfave/cli/v2"
)

// Type erased stream decoder
type streamDecoder func([]byte) (interface{}, error)

var FLAGS = []cli.Flag{
	FLAG_CONN_TYPE,
	FLAG_HOST,
}

var COMMANDS = []*cli.Command{
	CommandUpload,
	CommandDownload,
	CommandSession,
	CommandMetrics,
	CommandCaptures,
	CommandCapture,
	CommandEvents,
	CommandEvent,
	CommandProperties,
	CommandWalk,
}

func main() {
	app := &cli.App{
		Name:     "telemetry",
		Flags:    FLAGS,
		Commands: COMMANDS,
	}
	app.RunAndExitOnError()
}
