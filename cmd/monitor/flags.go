package main

import "github.com/urfave/cli/v2"

var (
	FLAG_PROMETHEUS_ADDRESS = &cli.StringFlag{
		Name:    "prometheus-address",
		Aliases: []string{"prometheus"},
		Usage:   "address of the monitor",
		EnvVars: []string{"MONITOR_PROMETHEUS_ADDRESS"},
		Value:   "localhost:2112",
	}

	FLAG_INFLUXDB_ADDRESS = &cli.StringFlag{
		Name:    "influxdb-address",
		Usage:   "address for influxdb",
		EnvVars: []string{"MONITOR_INFLUXDB_ADDRESS"},
		Value:   "localhost:8086",
	}

	FLAG_INFLUXDB_TOKEN = &cli.StringFlag{
		Name:     "influxdb-token",
		Usage:    "token for influxdb",
		EnvVars:  []string{"MONITOR_INFLUXDB_TOKEN"},
		Required: true,
	}

	FLAG_INFLUXDB_ORG = &cli.StringFlag{
		Name:     "influxdb-org",
		Usage:    "org for influxdb",
		EnvVars:  []string{"MONITOR_INFLUXDB_ORG"},
		Required: true,
	}

	FLAG_INFLUXDB_BUCKET = &cli.StringFlag{
		Name:     "influxdb-bucket",
		Usage:    "bucket for influxdb",
		EnvVars:  []string{"MONITOR_INFLUXDB_BUCKET"},
		Required: true,
	}

	FLAG_ADDRESS = &cli.StringFlag{
		Name:    "address",
		Aliases: []string{"addr"},
		Usage:   "listen address for the server",
		EnvVars: []string{"MONITOR_ADDRESS"},
		Value:   "localhost:4640",
	}

	FLAG_MAX_FAILED_ATTEMPTS = &cli.IntFlag{
		Name:    "max-failed-attemps",
		Usage:   "how many consecutive errors can happen while making requests to a peer before removing it",
		EnvVars: []string{"MONITOR_MAX_FAILED_ATTEMPS"},
	}

	FLAG_RETRY_INTERVAL = &cli.IntFlag{
		Name:    "retry-interval",
		Usage:   "how many seconds before retrying a request to a peer after a failure",
		EnvVars: []string{"MONITOR_RETRY_INTERVAL"},
	}

	FLAG_COLLECT_PERIOD = &cli.IntFlag{
		Name:    "collect-period",
		Usage:   "how many seconds between each telemetry request to a peer",
		EnvVars: []string{"MONITOR_COLLECT_PERIOD"},
	}

	FLAG_BANDWIDTH_PERIOD = &cli.IntFlag{
		Name:    "bandwidth-period",
		Usage:   "how many seconds between each bandwidth request to a peer",
		EnvVars: []string{"MONITOR_BANDWIDTH_PERIOD"},
	}

	FLAG_POSTGRES = &cli.StringFlag{
		Name:    "postgres-address",
		Usage:   "url for the database",
		EnvVars: []string{"MONITOR_POSTGRES"},
		Value:   "postgres://postgres@localhost/postgres?sslmode=disable",
	}
)
