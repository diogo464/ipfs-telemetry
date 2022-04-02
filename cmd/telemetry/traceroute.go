package main

import (
	"fmt"

	"git.d464.sh/adc/telemetry/pkg/traceroute"
	"github.com/urfave/cli/v2"
)

var CommandTraceRoute = &cli.Command{
	Name:    "traceroute",
	Aliases: []string{"tr"},
	Action:  actionTraceRoute,
}

func actionTraceRoute(c *cli.Context) error {
	result, err := traceroute.Trace(c.Args().First())
	if err != nil {
		return err
	}
	fmt.Println("PROVIDER =", result.Provider)
	fmt.Println(string(result.Output))
	return nil
}
