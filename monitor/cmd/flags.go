package main

import "github.com/urfave/cli/v2"

var (
	FLAG_PROMETHEUS_ADDRESS = &cli.StringFlag{
		Name:    "prometheus-address",
		Aliases: []string{"prometheus"},
		Usage:   "address of the monitor",
		EnvVars: []string{"PROMETHEUS_ADDRESS"},
		Value:   "0.0.0.0:9090",
	}

	FLAG_NATS_ENDPOINT = &cli.StringFlag{
		Name:    "nats-endpoint",
		Aliases: []string{"nats"},
		Usage:   "url of the nats server",
		EnvVars: []string{"NATS_ENDPOINT"},
		Value:   "nats://localhost:4222",
	}

	FLAG_MAX_FAILED_ATTEMPTS = &cli.IntFlag{
		Name:    "max-failed-attemps",
		Usage:   "how many consecutive errors can happen while making requests to a peer before removing it",
		EnvVars: []string{"MONITOR_MAX_FAILED_ATTEMPS"},
	}

	FLAG_RETRY_INTERVAL = &cli.DurationFlag{
		Name:    "retry-interval",
		Usage:   "how many seconds before retrying a request to a peer after a failure",
		EnvVars: []string{"MONITOR_RETRY_INTERVAL"},
	}

	FLAG_COLLECT_ENABLED = &cli.BoolFlag{
		Name:    "collect-enabled",
		EnvVars: []string{"MONITOR_COLLECT_ENABLED"},
		Value:   true,
	}

	FLAG_COLLECT_INTERVAL = &cli.DurationFlag{
		Name:    "collect-interval",
		Usage:   "how long between each telemetry request to a peer",
		EnvVars: []string{"MONITOR_COLLECT_INTERVAL"},
	}

	FLAG_COLLECT_TIMEOUT = &cli.DurationFlag{
		Name:    "collect-timeout",
		Usage:   "how long before a telemetry request times out and counts as an error",
		EnvVars: []string{"MONITOR_COLLECT_TIMEOUT"},
	}

	FLAG_BANDWIDTH_ENABLED = &cli.BoolFlag{
		Name:    "bandwidth-enabled",
		EnvVars: []string{"MONITOR_BANDWIDTH_ENABLED"},
		Value:   true,
	}

	FLAG_BANDWIDTH_INTERVAL = &cli.DurationFlag{
		Name:    "bandwidth-interval",
		Usage:   "how long between each bandwidth request to a peer",
		EnvVars: []string{"MONITOR_BANDWIDTH_INTERVAL"},
	}

	FLAG_BANDWIDTH_TIMEOUT = &cli.DurationFlag{
		Name:    "bandwidth-timeout",
		Usage:   "how long before a bandwidth request times out and counts as an error",
		EnvVars: []string{"MONITOR_BANDWIDTH_TIMEOUT"},
	}
)
