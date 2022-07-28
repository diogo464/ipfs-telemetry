package main

import (
	"context"
	"fmt"

	"git.d464.sh/uni/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandUpload = &cli.Command{
	Name:   "upload",
	Action: actionUpload,
}

func actionUpload(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	rate, err := client.Upload(context.Background(), telemetry.DEFAULT_PAYLOAD_SIZE)
	if err != nil {
		return err
	}
	fmt.Println("Upload rate:", rate/(1024*1024), "MB/s")
	return nil
}
