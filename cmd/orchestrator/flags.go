package main

import "github.com/urfave/cli/v2"

var (
	FLAG_PROBES = &cli.StringFlag{
		Name:    "probes",
		Usage:   "comma seperated list of addresses",
		EnvVars: []string{"ORCHESTRATOR_PROBES"},
	}

	FLAG_NUM_CIDS = &cli.IntFlag{
		Name:    "num-cids",
		EnvVars: []string{"ORCHESTRATOR_NUM_CIDS"},
		Value:   16,
	}

	FLAG_INFLUXDB_ADDRESS = &cli.StringFlag{
		Name:    "influxdb-address",
		Usage:   "address for influxdb",
		EnvVars: []string{"ORCHESTRATOR_INFLUXDB_ADDRESS"},
		Value:   "localhost:8086",
	}

	FLAG_INFLUXDB_TOKEN = &cli.StringFlag{
		Name:     "influxdb-token",
		Usage:    "token for influxdb",
		EnvVars:  []string{"ORCHESTRATOR_INFLUXDB_TOKEN"},
		Required: true,
	}

	FLAG_INFLUXDB_ORG = &cli.StringFlag{
		Name:     "influxdb-org",
		Usage:    "org for influxdb",
		EnvVars:  []string{"ORCHESTRATOR_INFLUXDB_ORG"},
		Required: true,
	}

	FLAG_INFLUXDB_BUCKET = &cli.StringFlag{
		Name:     "influxdb-bucket",
		Usage:    "bucket for influxdb",
		EnvVars:  []string{"ORCHESTRATOR_INFLUXDB_BUCKET"},
		Required: true,
	}
)
