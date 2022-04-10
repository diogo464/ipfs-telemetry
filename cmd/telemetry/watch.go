package main

import (
	"strings"
	"time"

	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"git.d464.sh/adc/telemetry/pkg/utils"
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

	snapshotTypes := strings.Split(c.String(FLAG_TYPE.Name), ",")
	if snapshotTypes[0] == "" {
		snapshotTypes = []string{}
	}

	var since uint64 = 0
	ticker := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-ticker.C:
			csnapshots := make(chan telemetry.SnapshotStreamItem)
			go func() {
				for item := range csnapshots {
					for _, s := range item.Snapshots {
						since = item.NextSeqN
						if len(snapshotTypes) == 0 || utils.SliceAny(snapshotTypes, func(t string) bool {
							return t == s.GetName()
						}) {
							printAsJson(s)
						}
					}
				}
			}()
			err := client.Snapshots(c.Context, since, csnapshots)
			if err != nil {
				return err
			}
		case <-c.Context.Done():
			break LOOP
		}
	}

	return nil
}
