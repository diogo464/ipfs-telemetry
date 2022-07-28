package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"git.d464.sh/uni/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/urfave/cli/v2"
)

var app = &cli.App{
	Name:   "node",
	Action: actionMain,
}

var EventMyEvent = telemetry.JsonEvent[MyEvent](telemetry.JsonEventDescriptor{
	Name:   "my_event",
	Period: time.Second,
})

type MyEvent struct {
	CurrentTime time.Time `json:"current_time"`
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func actionMain(c *cli.Context) error {
	h, err := createHost()
	if err != nil {
		return err
	}

	t, err := telemetry.NewService(
		h,
		telemetry.WithServiceDebug(true),
		telemetry.WithTcpListener("127.0.0.1:4000"),
		telemetry.WithServiceDefaultStreamOpts(telemetry.WithStreamActiveBufferLifetime(time.Second*5)),
	)
	if err != nil {
		return err
	}

	var PropConnCount = telemetry.JsonProperty(telemetry.JsonPropertyDescriptor{
		Name: "conn_count",
		Collect: func(ctx context.Context) (interface{}, error) {
			return struct {
				Count int `json:"count"`
			}{len(h.Network().Conns())}, nil
		},
	})

	t.RegisterCollector(EventMyEvent)
	t.RegisterProperty(PropConnCount)

	fmt.Println(h.Network().ListenAddresses())
	fmt.Println(h.ID())

	go func() {
		for {
			EventMyEvent.Publish(&MyEvent{time.Now()})
			time.Sleep(time.Second)
		}
	}()

	<-c.Context.Done()

	return nil
}

func createHost() (host.Host, error) {
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/3001"))
	if err != nil {
		return nil, err
	}
	return h, nil
}
