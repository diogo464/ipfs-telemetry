package main

import (
	"context"
	"fmt"
	"io"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/rle"
	"github.com/urfave/cli/v2"
)

var CommandProperty = &cli.Command{
	Name:        "property",
	Description: "Get property",
	Action:      actionProperty,
	Flags:       []cli.Flag{FLAG_NDJSON},
}

var PropertyHandlers = map[string]func(context.Context, io.Reader, bool) error{
	datapoint.ProviderRecordsName: func(ctx context.Context, r io.Reader, ndjson bool) error {
		for {
			recordBin, err := rle.Read(r)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			record, err := datapoint.ProviderRecordDeserialize(recordBin)
			if err != nil {
				return err
			}

			printAsJson(record, ndjson)
		}
		return nil
	},
}

func actionProperty(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	propertyName := c.Args().First()
	handler, ok := PropertyHandlers[propertyName]
	if !ok {
		return fmt.Errorf("no handler for property %s", propertyName)
	}

	reader, err := client.Property(c.Context, propertyName)
	if err != nil {
		return err
	}

	return handler(c.Context, reader, c.Bool(FLAG_NDJSON.Name))
}

func getPropertyDescriptorByName(ctx context.Context, c *telemetry.Client, name string) (*telemetry.StreamDescriptor, error) {
	descriptors, err := c.AvailableStreams(ctx)
	if err != nil {
		return nil, err
	}

	for _, desc := range descriptors {
		if desc.Name == name {
			return &desc, nil
		}
	}

	return nil, fmt.Errorf("stream %s not found", name)
}
