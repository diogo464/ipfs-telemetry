package main

import "github.com/urfave/cli/v2"

var (
	FLAG_CONCURRENCY = &cli.IntFlag{
		Name:    "concurrency",
		Usage:   "how many peers to request at the same time",
		EnvVars: []string{"WALK_CONCURRENCY"},
	}

	FLAG_CONNECT_TIMEOUT = &cli.DurationFlag{
		Name:    "connect-timeout",
		Usage:   "how long before a connection attempt times out",
		EnvVars: []string{"WALK_CONNECT_TIMEOUT"},
	}

	FLAG_REQUEST_TIMEOUT = &cli.DurationFlag{
		Name:    "request-timeout",
		Usage:   "how long before a request times out",
		EnvVars: []string{"WALK_REQUEST_TIMEOUT"},
	}

	FLAG_INTERVAL = &cli.DurationFlag{
		Name:    "interval",
		Usage:   "how long to wait between each peer request",
		EnvVars: []string{"WALK_INTERVAL"},
	}

	FLAG_TCP = &cli.BoolFlag{
		Name:    "tcp",
		Usage:   "use tcp to connect to peers",
		EnvVars: []string{"WALK_TCP"},
	}

	FLAG_UDP = &cli.BoolFlag{
		Name:    "udp",
		Usage:   "use udp to connect to peers",
		EnvVars: []string{"WALK_UDP"},
	}

	FLAG_IPV4 = &cli.BoolFlag{
		Name:    "ipv4",
		Usage:   "use ipv4 to connect to peers",
		EnvVars: []string{"WALK_IPV4"},
	}

	FLAG_IPV6 = &cli.BoolFlag{
		Name:    "ipv6",
		Usage:   "use ipv6 to connect to peers",
		EnvVars: []string{"WALK_IPV6"},
	}

	FLAG_OUTPUT = &cli.StringFlag{
		Name:    "output",
		Usage:   "output file",
		EnvVars: []string{"WALK_OUTPUT"},
	}

	FLAG_COMPRESS = &cli.BoolFlag{
		Name:    "compress",
		Usage:   "gzip compress the output",
		EnvVars: []string{"WALK_COMPRESS"},
	}
)
