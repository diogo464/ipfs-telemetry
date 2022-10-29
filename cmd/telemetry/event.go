package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandEvent = &cli.Command{
	Name:        "event",
	Description: "Get events",
	Action:      actionEvent,
}

func actionEvent(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	events, err := client.GetEvent(c.Context, c.Args().First(), 0)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	for _, ev := range events {
		buf := &bytes.Buffer{}
		buf.Reset()
		if err := json.Indent(buf, ev.Data, "", "  "); err != nil {
			return err
		}
		fmt.Println(string(buf.Bytes()))
	}

	return nil
}
