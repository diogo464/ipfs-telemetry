package main

import (
	"fmt"

	"github.com/diogo464/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandStream = &cli.Command{
	Name:        "stream",
	Description: "Stream data",
	Action:      actionStream,
}

func actionStream(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	streamName := c.Args().Get(0)
	segments := make(chan telemetry.StreamSegment)
	go client.StreamSegments(c.Context, streamName, 0, segments)
	for segment := range segments {
		msgs, err := telemetry.StreamSegmentDecode(telemetry.Int64StreamDecoder, segment)
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			fmt.Println(msg.Timestamp, streamName, "=", msg.Value)
		}
	}

	return nil
}
