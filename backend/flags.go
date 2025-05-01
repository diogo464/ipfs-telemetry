package backend

import (
	"github.com/urfave/cli/v2"
)

var (
	Flag_PrometheusAddress = &cli.StringFlag{
		Name:    "prometheus-address",
		Aliases: []string{"prometheus"},
		Usage:   "address used to expose prometheus metrics over http",
		EnvVars: []string{"PROMETHEUS_ADDRESS"},
		Value:   "0.0.0.0:9090",
	}

	Flag_NatsUrl = &cli.StringFlag{
		Name:    "nats-url",
		Usage:   "nats server url",
		EnvVars: []string{"NATS_URL"},
		Value:   "nats://localhost:4222",
	}

	Flag_VmUrl = &cli.StringFlag{
		Name:    "vm-url",
		Usage:   "VictoriaMetrics server url",
		EnvVars: []string{"VM_URL"},
		Value:   "http://localhost:8428",
	}

	Flag_PostgresUrl = &cli.StringFlag{
		Name:    "postgres-url",
		Usage:   "PostgreSQL url",
		EnvVars: []string{"POSTGRES_URL"},
		Value:   "postgres://postgres@localhost:5432/postgres",
	}
)
