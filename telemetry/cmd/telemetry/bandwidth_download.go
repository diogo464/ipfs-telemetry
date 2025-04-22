package main

import (
	"context"
	"fmt"

	"github.com/diogo464/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandDownload = &cli.Command{
	Name:   "download",
	Action: actionDownload,
}

func actionDownload(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	rate, err := client.Download(context.Background(), telemetry.DEFAULT_BANDWIDTH_PAYLOAD_SIZE)
	if err != nil {
		return err
	}
	fmt.Println("Download rate:", rate/(1024*1024), "MB/s")
	return nil
}
