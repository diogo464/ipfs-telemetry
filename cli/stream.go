package cli

import (
	"encoding/json"
	"fmt"

	"github.com/diogo464/telemetry"
	"github.com/urfave/cli/v2"
)

var fLAG_NDJSON = &cli.BoolFlag{Name: "ndjson", Value: false, Usage: "output in ndjson format"}

var commandStream = &cli.Command{
	Name:        "stream",
	Description: "stream data from a node",
	Action:      actionStream,
	Flags: []cli.Flag{
		fLAG_NDJSON,
	},
}

func defaultJsonDecoder(data []byte) (interface{}, error) {
	m, err := telemetry.JsonObjStreamDecoder(data)
	if err != nil {
		return nil, err
	}
	return m, nil
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

	var decoder StreamDecoder = nil
	if descriptor.Encoding == "json" {
		decoder = defaultJsonDecoder
	} else {
		decoder = registeredStreamDecoders[descriptor.Name]
	}
	if decoder == nil {
		return fmt.Errorf("No converter registered for encoding %s and name %s", descriptor.Encoding, descriptor.Name)
	}

	segments := make(chan telemetry.StreamSegment)
	go client.StreamSegments(c.Context, descriptor.Name, 0, segments)

	for segment := range segments {
		interfaces, err := telemetry.StreamSegmentDecode(telemetry.StreamDecoder[interface{}](decoder), segment)
		if err != nil {
			return err
		}
		for _, obj := range interfaces {
			if c.Bool(fLAG_NDJSON.Name) {
				marshaled, err := json.Marshal(obj)
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled))
			} else {
				marshaled, err := json.MarshalIndent(obj, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(marshaled))
			}
		}
	}

	return nil
}
