package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandEvents = &cli.Command{
	Name:        "events",
	Description: "List available events",
	Action:      actionEvents,
}

func actionEvents(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	events, err := client.GetAvailableEvents(c.Context)
	if err != nil {
		return err
	}

	for _, event := range events {
		fmt.Println(event.Name)
	}

	return nil
}
