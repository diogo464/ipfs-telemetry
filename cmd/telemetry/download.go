package main

import (
	"context"
	"fmt"

	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/urfave/cli"
)

var CommandDownload = cli.Command{
	Name:   "download",
	Action: actionDownload,
}

func actionDownload(c *cli.Context) error {
	client, err := clientFromAddr(c.Args()[0])
	if err != nil {
		return err
	}
	rate, err := client.Download(context.Background(), telemetry.DEFAULT_PAYLOAD_SIZE)
	if err != nil {
		return err
	}
	fmt.Println("Download rate:", rate/(1024*1024), "MB/s")
	return nil
}
