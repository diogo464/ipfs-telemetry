package crawler

import "github.com/urfave/cli/v2"

var (
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
