package main

import "github.com/urfave/cli/v2"

var (
	FLAG_PROMETHEUS_ADDRESS = &cli.StringFlag{
		Name:    "prometheus-address",
		Aliases: []string{"prometheus"},
		Usage:   "listen address for prometheus",
		EnvVars: []string{"PROBE_PROMETHEUS_ADDRESS"},
		Value:   "localhost:9090",
	}

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
