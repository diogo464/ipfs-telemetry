package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "monitor",
		Description: "collect telemetry from ipfs nodes",
		Action:      mainAction,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	return nil
}
