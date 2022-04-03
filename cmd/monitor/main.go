package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"git.d464.sh/adc/telemetry/pkg/monitor"
	pb "git.d464.sh/adc/telemetry/pkg/proto/monitor"
	"git.d464.sh/adc/telemetry/pkg/snapshot"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	listener, err := net.Listen("tcp", c.String(FLAG_ADDRESS.Name))
	if err != nil {
		return err
	}
	grpc_server := grpc.NewServer()

	//url := c.String(FLAG_DATABASE.Name)
	//db, err := sql.Open("postgres", url)
	//if err != nil {
	//	return err
	//}
	//defer db.Close()

	exporter := monitor.NewExporterFn(func(i peer.ID, s []snapshot.Snapshot) {
		fmt.Printf("Received %v snapshots\n", len(s))
	})

	server, err := monitor.NewMonitor(c.Context, exporter)
	if err != nil {
		return err
	}
	defer server.Close()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	pb.RegisterMonitorServer(grpc_server, server)
	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
	}()

	server.Run(c.Context)
	grpc_server.GracefulStop()

	return nil
}
