package main

import (
	"context"
	"fmt"

	"github.com/diogo464/ipfs_telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandStream = &cli.Command{
	Name:        "stream",
	Description: "Get available data from a stream",
	Action:      actionStream,
}

func actionStream(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	streamName := c.Args().First()
	decoder, ok := datapoint.Decoders[streamName]
	if !ok {
		return fmt.Errorf("no decoder for stream %s", streamName)
	}

	ch := make(chan telemetry.StreamSegment)
	go client.StreamSegments(c.Context, streamName, 0, ch)
	for seg := range ch {
		objs, err := telemetry.StreamSegmentDecode(decoder, seg)
		if err != nil {
			return err
		}
		for _, obj := range objs {
			printAsJson(obj, c.Bool(FLAG_NDJSON.Name))
		}
	}

	return nil
}

func getStreamDescriptorByName(ctx context.Context, c *telemetry.Client, name string) (*telemetry.StreamDescriptor, error) {
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
