package main

import (
	"context"
	"fmt"

	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandDownload = &cli.Command{
	Name:   "download",
	Action: actionDownload,
}

func actionDownload(c *cli.Context) error {
	client, err := clientFromAddr(c.Args().Slice()[0])
	if err != nil {
		return err
	}
	defer client.Close()

	rate, err := client.Download(context.Background(), telemetry.DEFAULT_PAYLOAD_SIZE)
	if err != nil {
		return err
	}
	fmt.Println("Download rate:", rate/(1024*1024), "MB/s")
	return nil
}
