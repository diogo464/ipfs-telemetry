package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

var CommandProviderRecords = &cli.Command{
	Name:    "provider-records",
	Aliases: []string{},
	Action:  actionProviderRecords,
}

func actionProviderRecords(c *cli.Context) error {
	client, err := clientFromAddr(c.Args().Slice()[0])
	if err != nil {
		return err
	}
	defer client.Close()

	crecords, err := client.ProviderRecords(context.Background())
	if err != nil {
		return err
	}
	for record := range crecords {
		printAsJson(record)
	}
	return nil
}
