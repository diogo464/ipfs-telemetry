package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandSnapshots = &cli.Command{
	Name:        "snapshots",
	Description: "List available snapshots",
	Action:      actionSnapshots,
}

func actionSnapshots(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	snapshots, err := client.GetAvailableSnapshots(c.Context)
	if err != nil {
		return err
	}

	for _, event := range snapshots {
		fmt.Println(event.Name)
	}

	return nil
}
