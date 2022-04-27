package main

import "github.com/urfave/cli/v2"

var (
	FLAG_ADDRESS = &cli.StringFlag{
		Name:    "address",
		Aliases: []string{"addr"},
		Usage:   "listen address for the server",
		EnvVars: []string{"PROBE_ADDRESS"},
		Value:   "0.0.0.0:4640",
	}

	FLAG_NAME = &cli.StringFlag{
		Name:     "name",
		EnvVars:  []string{"PROBE_NAME"},
		Required: true,
	}
)
