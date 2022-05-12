package main

import (
	"context"
	"fmt"
	"time"

	"github.com/diogo464/telemetry/pkg/crawler"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var FLAG_CLOSE = &cli.BoolFlag{
	Name:  "close",
	Usage: "close the stream there are no more peers",
	Value: false,
}

var SubscribeCommand = &cli.Command{
	Name:        "subscribe",
	Aliases:     []string{},
	Description: "subscribe to the crawler's output",
	Action:      subscribeAction,
	Flags: []cli.Flag{
		FLAG_CLOSE,
	},
}

func subscribeAction(c *cli.Context) error {
	conn, err := grpc.Dial(c.String(FLAG_ADDRESS.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client := crawler.NewClient(conn)
	csubscribe := make(chan peer.ID)
	ctx, cancel := context.WithCancel(c.Context)
	timer := time.NewTimer(time.Millisecond * 200)
	go func() {
		<-timer.C
		cancel()
	}()
	go func() {
		for p := range csubscribe {
			fmt.Println(p.Pretty())
			timer.Reset(time.Millisecond * 200)
		}
	}()

	err = client.Subscribe(ctx, csubscribe)
	if context.Canceled.Error() == status.Convert(err).Message() {
		return nil
	} else {
		return err
	}
}
