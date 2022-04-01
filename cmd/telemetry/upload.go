package main

import (
	"context"
	"fmt"

	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/urfave/cli"
)

var CommandUpload = cli.Command{
	Name:   "upload",
	Action: actionUpload,
}

func actionUpload(c *cli.Context) error {
	client, err := clientFromAddr(c.Args()[0])
	if err != nil {
		return err
	}
	rate, err := client.Upload(context.Background(), telemetry.DEFAULT_PAYLOAD_SIZE)
	if err != nil {
		return err
	}
	fmt.Println("Upload rate:", rate/(1024*1024), "MB/s")
	return nil
}
