package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"git.d464.sh/adc/telemetry/pkg/monitor"
	pb "git.d464.sh/adc/telemetry/pkg/proto/monitor"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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

	client := influxdb2.NewClient("http://localhost:8086", "1tzS8zED9lUkUXniy4sFwq3193_5vPnXLcTPsgpwe4gi6oP_mDc1sicifSR00P8i8HZnNz0FNsv8-tRkH_-Pcw==")
	defer client.Close()
	writeAPI := client.WriteAPI("adc", "telemetry")
	exporter := NewInfluxExporter(writeAPI)
	defer exporter.Close()

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
