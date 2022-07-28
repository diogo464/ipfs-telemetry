package cli

import "github.com/urfave/cli/v2"

var (
	fLAG_CONN_TYPE = &cli.StringFlag{
		Name:     "conn-type",
		Usage:    "Connection type. 'tcp' or 'libp2p'",
		Value:    "tcp",
		Required: false,
	}

	fLAG_HOST = &cli.StringFlag{
		Name:    "host",
		Usage:   "Host to connect to (e.g. 'localhost:8080' if conn-type is tcp)",
		Value:   "localhost:4000",
		EnvVars: []string{"TELEMETRY_HOST"},
	}
)
