package main

import "github.com/urfave/cli/v2"

var (
	FLAG_PROMETHEUS_ADDRESS = &cli.StringFlag{
		Name:    "prometheus-address",
		Aliases: []string{"prometheus"},
		Usage:   "address of the monitor",
		EnvVars: []string{"CRAWLER_PROMETHEUS_ADDRESS"},
		Value:   "localhost:2113",
	}

	FLAG_INFLUXDB_ADDRESS = &cli.StringFlag{
		Name:    "influxdb-address",
		Usage:   "address for influxdb",
		EnvVars: []string{"CRAWLER_INFLUXDB_ADDRESS"},
		Value:   "localhost:8086",
	}

	FLAG_INFLUXDB_TOKEN = &cli.StringFlag{
		Name:     "influxdb-token",
		Usage:    "token for influxdb",
		EnvVars:  []string{"CRAWLER_INFLUXDB_TOKEN"},
		Required: true,
	}

	FLAG_INFLUXDB_ORG = &cli.StringFlag{
		Name:     "influxdb-org",
		Usage:    "org for influxdb",
		EnvVars:  []string{"CRAWLER_INFLUXDB_ORG"},
		Required: true,
	}

	FLAG_INFLUXDB_BUCKET = &cli.StringFlag{
		Name:     "influxdb-bucket",
		Usage:    "bucket for influxdb",
		EnvVars:  []string{"CRAWLER_INFLUXDB_BUCKET"},
		Required: true,
	}

	FLAG_ADDRESS = &cli.StringFlag{
		Name:    "address",
		Aliases: []string{"addr"},
		Usage:   "listen address for the crawler",
		EnvVars: []string{"CRAWLER_ADDRESS"},
		Value:   "localhost:4641",
	}
)
