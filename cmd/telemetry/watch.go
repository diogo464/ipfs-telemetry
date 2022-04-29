package main

import (
	"strings"
	"time"

	"github.com/diogo464/telemetry/pkg/telemetry"
	"github.com/diogo464/telemetry/pkg/utils"
	"github.com/urfave/cli/v2"
)

var FLAG_TYPE = &cli.StringFlag{Name: "type"}

var CommandWatch = &cli.Command{
	Name:   "watch",
	Action: actionWatch,
	Flags: []cli.Flag{
		FLAG_TYPE,
	},
}

func actionWatch(c *cli.Context) error {
	client, err := clientFromAddr(c.Args().First())
	if err != nil {
		return err
	}
	defer client.Close()

	datapointTypes := strings.Split(c.String(FLAG_TYPE.Name), ",")
	if datapointTypes[0] == "" {
		datapointTypes = []string{}
	}

	var since uint64 = 0
	ticker := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-ticker.C:
			cdatapoints := make(chan telemetry.DatapointStreamItem)
			go func() {
				for item := range cdatapoints {
					for _, s := range item.Datapoints {
						since = item.NextSeqN
						if len(datapointTypes) == 0 || utils.SliceAny(datapointTypes, func(t string) bool {
							return t == s.GetName()
						}) {
							printAsJson(s)
						}
					}
				}
			}()
			err := client.Datapoints(c.Context, since, cdatapoints)
			if err != nil {
				return err
			}
		case <-c.Context.Done():
			break LOOP
		}
	}

	return nil
}
