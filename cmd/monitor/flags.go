package main

import "github.com/urfave/cli/v2"

var (
	FLAG_DATABASE = &cli.StringFlag{
		Name:    "database",
		Aliases: []string{"d"},
		Usage:   "url for the database",
		EnvVars: []string{"MONITOR_DATABASE"},
		Value:   "postgres://postgres@localhost/postgres?sslmode=disable",
	}

	FLAG_ADDRESS = &cli.StringFlag{
		Name:    "address",
		Aliases: []string{"addr"},
		Usage:   "listen address for the server",
		EnvVars: []string{"MONITOR_ADDRESS"},
		Value:   "localhost:5000",
	}
)
