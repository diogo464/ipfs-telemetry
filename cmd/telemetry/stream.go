package main

import (
	"fmt"

	"git.d464.sh/uni/telemetry"
	"github.com/urfave/cli/v2"
)

var FLAG_NDJSON = &cli.BoolFlag{Name: "ndjson", Value: false, Usage: "output in ndjson format"}

var CommandStream = &cli.Command{
	Name:        "stream",
	Description: "stream data from a node",
	Action:      actionStream,
	Flags: []cli.Flag{
		FLAG_NDJSON,
	},
}

func actionStream(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	streams, err := client.AvailableStreams(c.Context)
	if err != nil {
		return err
	}

	var descriptor *telemetry.StreamDescriptor = nil
	for _, stream := range streams {
		if stream.Name == c.Args().First() {
			descriptor = &stream
			break
		}
	}
	if descriptor == nil {
		return fmt.Errorf("Stream not found")
	}
	if descriptor.Encoding != "json" {
		return fmt.Errorf("Stream encoding is not json")
	}

	segments := make(chan telemetry.StreamSegment)
	go client.StreamSegments(c.Context, descriptor.Name, 0, segments)
	decoder := telemetry.JsonPrettyStreamDecoder
	if c.Bool(FLAG_NDJSON.Name) {
		decoder = telemetry.JsonStreamDecoder
	}

	for segment := range segments {
		pretty, err := telemetry.StreamSegmentDecode(decoder, segment)
		if err != nil {
			fmt.Println("data:", string(segment.Data))
			return err
		}
		for _, elem := range pretty {
			fmt.Println(elem)
		}
	}

	return nil
}
