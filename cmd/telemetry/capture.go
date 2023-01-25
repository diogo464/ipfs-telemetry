package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/diogo464/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandCapture = &cli.Command{
	Name:        "capture",
	Description: "Get the latest data from a capture",
	Action:      actionCapture,
}

func actionCapture(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	id, err := strconv.Atoi(c.Args().First())
	if err != nil {
		return err
	}
	captures, err := client.GetCapture(c.Context, telemetry.StreamId(id))
	if err != nil {
		return err
	}
	if len(captures) == 0 {
		return nil
	}

	capture := captures[len(captures)-1]
	buf := &bytes.Buffer{}
	buf.Reset()
	if err := json.Indent(buf, capture.Data, "", "  "); err != nil {
		return err
	}
	fmt.Println(buf.String())

	return nil
}
