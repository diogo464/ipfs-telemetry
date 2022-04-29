package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/diogo464/telemetry/pkg/monitor"
	pb "github.com/diogo464/telemetry/pkg/proto/monitor"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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
			FLAG_PROMETHEUS_ADDRESS,
			FLAG_INFLUXDB_ADDRESS,
			FLAG_INFLUXDB_TOKEN,
			FLAG_INFLUXDB_ORG,
			FLAG_INFLUXDB_BUCKET,
			FLAG_ADDRESS,
			FLAG_POSTGRES,
		},
	}

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

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

	client := influxdb2.NewClient(c.String(FLAG_INFLUXDB_ADDRESS.Name), c.String(FLAG_INFLUXDB_TOKEN.Name))
	defer client.Close()
	writeAPI := client.WriteAPI(c.String(FLAG_INFLUXDB_ORG.Name), c.String(FLAG_INFLUXDB_BUCKET.Name))
	exporter := NewInfluxExporter(writeAPI)
	defer exporter.Close()

	server, err := monitor.NewMonitor(c.Context, exporter, monitor.WithBandwidthPeriod(time.Hour))
	if err != nil {
		return err
	}
	defer server.Close()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(c.String(FLAG_PROMETHEUS_ADDRESS.Name), nil))
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
