package main

import (
	"context"

	"github.com/urfave/cli"
)

var CommandSystemInfo = cli.Command{
	Name:    "system-info",
	Aliases: []string{"info"},
	Action:  actionSystemInfo,
}

func actionSystemInfo(c *cli.Context) error {
	client, err := clientFromAddr(c.Args()[0])
	if err != nil {
		return err
	}
	info, err := client.SystemInfo(context.Background())
	if err != nil {
		return err
	}
	printAsJson(info)
	return nil
}
