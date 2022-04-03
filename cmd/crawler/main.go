package main

import (
	"context"
	"fmt"
	"os"

	"git.d464.sh/adc/telemetry/pkg/crawler"
	"git.d464.sh/adc/telemetry/pkg/monitor"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

var FLAG_MONITOR_ADDRESS = &cli.StringFlag{
	Name:    "monitor-address",
	Aliases: []string{"monitor"},
	Usage:   "address of the monitor",
	EnvVars: []string{"CRAWLER_MONITOR_ADDRESS"},
	Value:   "localhost:5000",
}

func main() {
	app := &cli.App{
		Name:        "crawler",
		Description: "discovery peers supporting the telemetry protocol",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_MONITOR_ADDRESS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

type eventHandler struct {
	h      host.Host
	client monitor.Client
}

// craweler.EventHandler impl
func (e *eventHandler) OnConnect(p peer.ID) error {
	fmt.Println("CONNECTED")
	if protocols, err := e.h.Peerstore().GetProtocols(p); err == nil {
		for _, protocol := range protocols {
			if protocol == telemetry.ID_TELEMETRY {
				e.client.Discover(context.TODO(), p)
				break
			}
		}
	}
	return nil
}
func (e *eventHandler) OnFinish(p peer.ID, addrs []peer.AddrInfo) error { return nil }
func (e *eventHandler) OnFail(p peer.ID, err error) error               { return nil }

func mainAction(c *cli.Context) error {
	fmt.Println("connecting to monitior...")
	conn, err := grpc.Dial(c.String(FLAG_MONITOR_ADDRESS.Name), grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := monitor.NewClient(conn)
	fmt.Println("connected to monitor")

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return err
	}

	handler := &eventHandler{h, client}
	crwlr, err := crawler.NewCrawler(h, handler)
	if err != nil {
		return err
	}

	fmt.Println("starting crawl")
	for {
		err = crwlr.Crawl(c.Context)
		if err != nil {
			if err == context.Canceled {
				return nil
			} else {
				return err
			}
		}
	}
}
