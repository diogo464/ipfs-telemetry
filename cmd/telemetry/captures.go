package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandCaptures = &cli.Command{
	Name:        "captures",
	Description: "List available captures",
	Action:      actionCaptures,
}

func actionCaptures(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	descriptors, err := client.GetCaptureDescriptors(c.Context)
	if err != nil {
		return err
	}

	for _, desc := range descriptors {
		fmt.Println(desc.Name)
		fmt.Println("\t", desc.Description)
	}

	return nil
}
