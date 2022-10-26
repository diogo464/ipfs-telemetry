package main

import (
	"fmt"
	"sort"

	"github.com/urfave/cli/v2"
)

var CommandMetrics = &cli.Command{
	Name:        "metrics",
	Description: "Display latest metrics",
	Action:      actionMetrics,
}

func actionMetrics(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	response, err := client.GetMetrics(c.Context, 0)
	if err != nil {
		return err
	}
	if len(response.Observations) == 0 {
		return nil
	}

	observations := response.Observations[len(response.Observations)-1]
	keys := make([]string, 0, len(observations.Metrics))
	for key := range observations.Metrics {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Println(observations.Timestamp)
	for _, key := range keys {
		fmt.Println(key, "=", observations.Metrics[key])
	}
	return nil
}
