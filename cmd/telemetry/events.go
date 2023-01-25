package main

import (
	"fmt"
	"strconv"

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

	descriptors, err := client.GetEventDescriptors(c.Context)
	if err != nil {
		return err
	}

	for _, desc := range descriptors {
		fmt.Println(desc.Name + " - " + strconv.Itoa(int(desc.StreamId)))
		fmt.Println("\t", desc.Description)
	}

	return nil
}
