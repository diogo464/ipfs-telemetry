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

	FLAG_NATS_URL = &cli.StringFlag{
		Name:    "nats-url",
		Aliases: []string{"nats"},
		Usage:   "url of the nats server",
		EnvVars: []string{"NATS_URL"},
		Value:   "nats://nats:4222",
	}

	FLAG_NATS_SUBJECT = &cli.StringFlag{
		Name:    "nats-subject",
		Usage:   "nats subject to publish to",
		EnvVars: []string{"NATS_SUBJECT"},
		Value:   "discovery",
	}

	FLAG_OUTPUT = &cli.StringFlag{
		Name:    "output",
		Usage:   "crawler output method. can be one of 'http' or 'nats'",
		EnvVars: []string{"CRAWLER_OUTPUT"},
		Value:   "nats",
	}

	FLAG_HTTP_URL = &cli.StringFlag{
		Name:    "http-url",
		Usage:   "http api url. used when output mode is 'http'",
		EnvVars: []string{"CRAWLER_HTTP_URL"},
		Value:   "ipfs-telemetry.org/api/v1",
	}

	FLAG_CONCURRENCY = &cli.IntFlag{
		Name:    "concurrency",
		Usage:   "how many peers to request at the same time",
		EnvVars: []string{"CRAWLER_CONCURRENCY"},
	}

	FLAG_CONNECT_TIMEOUT = &cli.DurationFlag{
		Name:    "connect-timeout",
		Usage:   "how long before a connection attempt times out",
		EnvVars: []string{"CRAWLER_CONNECT_TIMEOUT"},
	}

	FLAG_REQUEST_TIMEOUT = &cli.DurationFlag{
		Name:    "request-timeout",
		Usage:   "how long before a request times out",
		EnvVars: []string{"CRAWLER_REQUEST_TIMEOUT"},
	}

	FLAG_INTERVAL = &cli.DurationFlag{
		Name:    "interval",
		Usage:   "how long to wait between each peer request",
		EnvVars: []string{"CRAWLER_INTERVAL"},
	}
)
