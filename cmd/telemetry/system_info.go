package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

var CommandSystemInfo = &cli.Command{
	Name:    "system-info",
	Aliases: []string{"info"},
	Action:  actionSystemInfo,
}

func actionSystemInfo(c *cli.Context) error {
	client, err := clientFromAddr(c.Args().Slice()[0])
	if err != nil {
		return err
	}
	defer client.Close()

	info, err := client.SystemInfo(context.Background())
	if err != nil {
		return err
	}
	printAsJson(info)
	return nil
}
