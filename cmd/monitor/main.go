package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/monitor"
	"github.com/ipfs/go-datastore"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func main() {
	app := &cli.App{
		Name:        "monitor",
		Description: "collect telemetry from ipfs nodes",
		Action:      mainAction,
		Commands:    []*cli.Command{DiscoverCommand},
		Flags: []cli.Flag{
			FLAG_DATABASE,
			FLAG_ADDRESS,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	h, err := createHost()
	if err != nil {
		return err
	}
	defer h.Close()

	listener, err := net.Listen("tcp", c.String(FLAG_ADDRESS.Name))
	if err != nil {
		return err
	}
	grpc_server := grpc.NewServer()

	url := c.String(FLAG_DATABASE.Name)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return err
	}
	defer db.Close()

	server, err := NewMonitor(h, db)
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

	//go func() {
	//	http.Handle("/metrics", promhttp.Handler())
	//	http.ListenAndServe(":2112", nil)
	//}()

	go func() {
		<-c.Context.Done()
		fmt.Println("CLI CONTEXT TERMINATED")
	}()

	pb.RegisterMonitorServer(grpc_server, server)
	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
	}()

	server.StartMonitoring(c.Context)
	fmt.Println("Staring GRPC graceful stop")
	grpc_server.GracefulStop()
	fmt.Println("GRPC stopped")

	return nil
}

func createHost() (host.Host, error) {
	return libp2p.New(
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
}
