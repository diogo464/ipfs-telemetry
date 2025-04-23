package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandDescriptors = &cli.Command{
	Name:        "descriptors",
	Description: "List all descriptors",
	Action:      actionDescriptors,
}

func actionDescriptors(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	descriptors, err := client.GetEventDescriptors(c.Context)
	if err != nil {
		return err
	}

	for _, d := range descriptors {
		fmt.Println("EventId: ", d.EventId)
		fmt.Println("\tScope: ", d.Scope.Name)
		fmt.Println("\tVersion: ", d.Scope.Version)
		fmt.Println("\tName: ", d.Name)
		fmt.Println("\tDescription: ", d.Description)
	}

	return nil
}
