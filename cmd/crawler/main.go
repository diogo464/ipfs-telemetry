package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"git.d464.sh/adc/telemetry/pkg/crawler"
	"git.d464.sh/adc/telemetry/pkg/monitor"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

var FLAG_PROMETHEUS_ADDRESS = &cli.StringFlag{
	Name:    "prometheus-address",
	Aliases: []string{"prometheus"},
	Usage:   "address of the monitor",
	EnvVars: []string{"CRAWLER_PROMETHEUS_ADDRESS"},
	Value:   "localhost:2113",
}

func main() {
	app := &cli.App{
		Name:        "crawler",
		Description: "discovery peers supporting the telemetry protocol",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_MONITOR_ADDRESS,
			FLAG_PROMETHEUS_ADDRESS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

type eventHandler struct {
	h      host.Host
	client monitor.Client

	peers map[peer.ID]struct{}
}

func newEventHandler(h host.Host, client monitor.Client) *eventHandler {
	return &eventHandler{
		h:      h,
		client: client,

		peers: make(map[peer.ID]struct{}),
	}
}

func (e *eventHandler) peerHasTelemetry(p peer.ID) (bool, error) {
	protocols, err := e.h.Peerstore().GetProtocols(p)
	if err != nil {
		return false, err
	}

	for _, protocol := range protocols {
		if protocol == telemetry.ID_TELEMETRY {
			return true, nil
		}
	}
	return false, nil
}

// craweler.EventHandler impl
func (e *eventHandler) OnConnect(p peer.ID) error {
	TotalCrawls.Add(1)

	hasTelemetry, err := e.peerHasTelemetry(p)
	if err != nil { // dont stop the crawl, just ignore this peer
		return nil
	}

	if _, ok := e.peers[p]; !ok {
		e.peers[p] = struct{}{}
		UniquePeers.Add(1)
		if hasTelemetry {
			UniquePeersTelemetry.Add(1)
		}
	}

	if hasTelemetry {
		if err := e.client.Discover(context.Background(), p); err != nil {
			return err
		}
	}

	return nil
}
func (e *eventHandler) OnFinish(p peer.ID, addrs []peer.AddrInfo) error {
	SuccessfulCrawls.Add(1)
	return nil
}
func (e *eventHandler) OnFail(p peer.ID, err error) error {
	FailedCrawls.Add(1)
	return nil
}

func mainAction(c *cli.Context) error {
	conn, err := grpc.Dial(c.String(FLAG_MONITOR_ADDRESS.Name), grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := monitor.NewClient(conn)

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return err
	}

	handler := newEventHandler(h, client)
	crwlr, err := crawler.NewCrawler(h, handler)
	if err != nil {
		return err
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(c.String(FLAG_PROMETHEUS_ADDRESS.Name), nil))
	}()

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
