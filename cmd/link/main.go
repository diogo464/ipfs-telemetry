package main

import (
	"fmt"
	"os"
	"sync"
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

func mainAction(c *cli.Context) error {
	for {
		var monitor_conn *grpc.ClientConn
		var crawler_conn *grpc.ClientConn
		var mon monitor.Client
		var crw *crawler.Client
		var peers chan peer.ID
		var wg *sync.WaitGroup = new(sync.WaitGroup)
		var err error

		monitor_conn, err = grpc.Dial(c.String(FLAG_MONITOR.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			goto end
		}

		crawler_conn, err = grpc.Dial(c.String(FLAG_CRAWLER.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			goto end
		}

		wg.Add(1)
		mon = monitor.NewClient(monitor_conn)
		crw = crawler.NewClient(crawler_conn)
		peers = make(chan peer.ID)
		go func() {
			for p := range peers {
				if err = mon.Discover(c.Context, p); err != nil {
					fmt.Println(err)
					break
				} else {
					fmt.Println("Discovered", p)
				}
			}
			wg.Done()
		}()

		go func() {
			err = crw.Subscribe(c.Context, peers)
			wg.Done()
		}()

		wg.Wait()
	end:
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
}
