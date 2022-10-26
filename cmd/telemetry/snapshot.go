package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandSnapshot = &cli.Command{
	Name:        "snapshot",
	Description: "Get latest data from snapshot",
	Action:      actionSnapshot,
}

func actionSnapshot(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	response, err := client.GetSnapshot(c.Context, c.Args().First(), 0)
	if err != nil {
		return err
	}
	if len(response.Snapshots) == 0 {
		return nil
	}

	rsnapshot := response.Snapshots[len(response.Snapshots)-1]
	buf := &bytes.Buffer{}
	buf.Reset()
	if err := json.Indent(buf, rsnapshot.Data, "", "  "); err != nil {
		return err
	}
	fmt.Println(string(buf.Bytes()))

	return nil
}
