package server

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/ipfs/go-datastore"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
)

var ServerCommand *cli.Command = &cli.Command{
	Name:        "server",
	Description: "collect telemetry from ipfs nodes",
	Action:      serverAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "database",
			Aliases: []string{"d"},
			Usage:   "url for the database",
			EnvVars: []string{"MONITOR_DATABASE"},
			Value:   "postgres://postgres@localhost/postgres?sslmode=disable",
		},
	},
}

func serverAction(c *cli.Context) error {
	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			client := dht.NewDHTClient(context.TODO(), h, datastore.NewMapDatastore())
			if err := client.Bootstrap(context.TODO()); err != nil {
				return nil, err
			}

			var err error = nil
			var success bool = false
			for _, bootstrap := range dht.GetDefaultBootstrapPeerAddrInfos() {
				err = h.Connect(context.TODO(), bootstrap)
				if err == nil {
					success = true
				}
			}

			if success {
				client.RefreshRoutingTable()
				time.Sleep(time.Second * 2)
				return client, nil
			} else {
				return nil, err
			}
		}))
	if err != nil {
		return err
	}
	defer h.Close()

	url := c.String("database")
	db, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}
	defer db.Close()

	server, err := NewServer(h, db)
	if err != nil {
		return err
	}

	pids := make([]peer.ID, 0, c.Args().Len())
	for _, spid := range c.Args().Slice() {
		pid, err := peer.Decode(spid)
		if err != nil {
			return err
		}
		pids = append(pids, pid)
	}

	go func() {
		for _, pid := range pids {
			server.PeerDiscovered(pid)
		}
	}()

	go func() {
		for {
			TELEMETRY_REQUESTS_TOTAL.Add(1)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

	server.StartMonitoring(c.Context)

	return nil
}
