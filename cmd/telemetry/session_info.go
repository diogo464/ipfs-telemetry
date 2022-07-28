package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

var CommandSessionInfo = &cli.Command{
	Name:    "session-info",
	Aliases: []string{"session"},
	Action:  actionSessionInfo,
}

func actionSessionInfo(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	info, err := client.SessionInfo(context.Background())
	if err != nil {
		return err
	}
	printAsJson(info)
	return nil
}
