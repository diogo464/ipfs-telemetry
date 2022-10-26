package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandEvent = &cli.Command{
	Name:        "event",
	Description: "Stream event data",
	Action:      actionEvent,
}

func actionEvent(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	response, err := client.GetEvent(c.Context, c.Args().First(), 0)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	for _, revent := range response.Events {
		buf.Reset()
		if err := json.Indent(buf, revent.Data, "", "  "); err != nil {
			return err
		}
		fmt.Println(string(buf.Bytes()))
	}

	return nil
}
