package main

import (
	"fmt"
	"os"

	"git.d464.sh/adc/telemetry/pkg/crawler"
	"git.d464.sh/adc/telemetry/pkg/monitor"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
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

func mainAction(c *cli.Context) error {
	monitor_conn, err := grpc.Dial(c.String(FLAG_MONITOR.Name), grpc.WithInsecure())
	if err != nil {
		return err
	}

	crawler_conn, err := grpc.Dial(c.String(FLAG_CRAWLER.Name), grpc.WithInsecure())
	if err != nil {
		return err
	}

	mon := monitor.NewClient(monitor_conn)
	crw := crawler.NewClient(crawler_conn)
	peers := make(chan peer.ID)
	go func() {
		for p := range peers {
			if err := mon.Discover(c.Context, p); err != nil {
				fmt.Println(err)
			}
		}
	}()
	return crw.Subscribe(c.Context, peers)
}
