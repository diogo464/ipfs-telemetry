package main

import (
	"fmt"

	"github.com/diogo464/telemetry"
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

	descriptors, err := client.GetStreamDescriptors(c.Context)
	if err != nil {
		return err
	}

	for _, descriptor := range descriptors {
		fmt.Println("StreamID: ", descriptor.ID)
		switch d := descriptor.Type.(type) {
		case *telemetry.StreamTypeMetric:
			fmt.Println("\tType: Metrics")
		case *telemetry.StreamTypeEvent:
			fmt.Println("\tType: Event")
			fmt.Println("\tScope: ", d.Scope.Name)
			fmt.Println("\tVersion: ", d.Scope.Version)
			fmt.Println("\tName: ", d.Name)
			fmt.Println("\tDescription: ", d.Description)
		}
	}

	return nil
}

func createDescriptor(kind string, desc interface{}) interface{} {
	return struct {
		Kind       string
		Descriptor interface{}
	}{
		Kind:       kind,
		Descriptor: desc,
	}
}
