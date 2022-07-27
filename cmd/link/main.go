package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/diogo464/telemetry/pkg/crawler"
	"github.com/diogo464/telemetry/pkg/monitor"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var FLAG_MONITOR = &cli.StringFlag{
	Name:    "monitor",
	Usage:   "address of the monitor",
	EnvVars: []string{"LINK_MONITOR"},
	Value:   "localhost:4640",
}

var FLAG_CRAWLER = &cli.StringFlag{
	Name:    "crawler",
	Usage:   "address of the crawler",
	EnvVars: []string{"LINK_CRAWLER"},
	Value:   "localhost:4641",
}

func main() {
	app := &cli.App{
		Name:        "link",
		Description: "link a monitor and a crawler",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_MONITOR,
			FLAG_CRAWLER,
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
	fmt.Println("Finished")
}

func notifyMonitor(c *cli.Context) error {
	fmt.Println("connecting to monitor")
	monitor_conn, err := grpc.Dial(c.String(FLAG_MONITOR.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer monitor_conn.Close()

	fmt.Println("connecting to crawler")
	crawler_conn, err := grpc.Dial(c.String(FLAG_CRAWLER.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer crawler_conn.Close()

	fmt.Println("creating clients")
	mon := monitor.NewClient(monitor_conn)
	crw := crawler.NewClient(crawler_conn)
	peers := make(chan peer.ID)
	fmt.Println("subscribing to crawler")
	go func() {
		ctx, cancel := context.WithTimeout(c.Context, time.Second*5)
		defer cancel()
		if err := crw.Subscribe(ctx, peers); err != nil {
			fmt.Println(err)
		}
	}()

	for p := range peers {
		if err := mon.Discover(c.Context, p); err != nil {
			return err
		} else {
			fmt.Println("Discovered", p)
		}
	}

	return nil
}

func mainAction(c *cli.Context) error {
	for {
		if err := notifyMonitor(c); err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 5)
	}
}
