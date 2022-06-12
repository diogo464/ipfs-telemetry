package main

import (
	"fmt"
	"time"

	"github.com/diogo464/telemetry/pkg/datapoint"
	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

var FLAG_TYPE = &cli.StringFlag{Name: "type", Required: true}
var FLAG_SINCE = &cli.Int64Flag{Name: "since", Value: 0}

var CommandWatch = &cli.Command{
	Name:   "watch",
	Action: actionWatch,
	Flags: []cli.Flag{
		FLAG_TYPE,
		FLAG_SINCE,
	},
}

func actionWatch(c *cli.Context) error {
	client, err := clientFromAddr(c.Args().First())
	if err != nil {
		return err
	}
	defer client.Close()

	decoder, ok := datapoint.Decoders[c.String(FLAG_TYPE.Name)]
	if !ok {
		return fmt.Errorf("Unknown datapoint type")
	}

	var since uint32 = uint32(c.Int64(FLAG_SINCE.Name))
	ticker := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-ticker.C:
			segments, err := client.Stream(c.Context, since, c.String(FLAG_TYPE.Name))
			if err != nil {
				return err
			}

			for _, segment := range segments {
				if uint32(segment.SeqN+1) > since {
					since = uint32(segment.SeqN + 1)
				}

				dps, err := telemetry.StreamSegmentDecode(decoder, segment)
				if err != nil {
					return err
				}
				for _, dp := range dps {
					printAsJson(dp)
				}
			}
		case <-c.Context.Done():
			break LOOP
		}
	}

	return nil
}
