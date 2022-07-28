package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/diogo464/ipfs_telemetry/pkg/monitor"
	pb "github.com/diogo464/ipfs_telemetry/pkg/proto/monitor"
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
			FLAG_MAX_FAILED_ATTEMPTS,
			FLAG_RETRY_INTERVAL,
			FLAG_COLLECT_ENABLED,
			FLAG_COLLECT_PERIOD,
			FLAG_COLLECT_TIMEOUT,
			FLAG_BANDWIDTH_ENABLED,
			FLAG_BANDWIDTH_PERIOD,
			FLAG_BANDWIDTH_TIMEOUT,
			FLAG_PROVIDER_RECORDS_ENABLED,
			FLAG_PROVIDER_RECORDS_PERIOD,
			FLAG_PROVIDER_RECORDS_TIMEOUT,
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

	monitorOptions := make([]monitor.Option, 0)

	if c.IsSet(FLAG_MAX_FAILED_ATTEMPTS.Name) {
		monitorOptions = append(monitorOptions, monitor.WithMaxFailedAttempts(c.Int(FLAG_MAX_FAILED_ATTEMPTS.Name)))
	}

	if c.IsSet(FLAG_RETRY_INTERVAL.Name) {
		monitorOptions = append(monitorOptions, monitor.WithRetryInterval(time.Second*time.Duration(c.Int(FLAG_RETRY_INTERVAL.Name))))
	}

	monitorOptions = append(monitorOptions, monitor.WithCollectEnabled(c.Bool(FLAG_COLLECT_ENABLED.Name)))

	if c.IsSet(FLAG_COLLECT_PERIOD.Name) {
		monitorOptions = append(monitorOptions, monitor.WithCollectPeriod(time.Second*time.Duration(c.Int(FLAG_COLLECT_PERIOD.Name))))
	}

	if c.IsSet(FLAG_COLLECT_TIMEOUT.Name) {
		monitorOptions = append(monitorOptions, monitor.WithCollectTimeout(time.Second*time.Duration(c.Int(FLAG_COLLECT_TIMEOUT.Name))))
	}

	monitorOptions = append(monitorOptions, monitor.WithBandwidthEnabled(c.Bool(FLAG_BANDWIDTH_ENABLED.Name)))

	if c.IsSet(FLAG_BANDWIDTH_PERIOD.Name) {
		monitorOptions = append(monitorOptions, monitor.WithBandwidthPeriod(time.Second*time.Duration(c.Int(FLAG_BANDWIDTH_PERIOD.Name))))
	}

	if c.IsSet(FLAG_BANDWIDTH_TIMEOUT.Name) {
		monitorOptions = append(monitorOptions, monitor.WithBandwidthTimeout(time.Second*time.Duration(c.Int(FLAG_BANDWIDTH_TIMEOUT.Name))))
	}

	monitorOptions = append(monitorOptions, monitor.WithProviderRecordsEnabled(c.Bool(FLAG_PROVIDER_RECORDS_ENABLED.Name)))

	if c.IsSet(FLAG_PROVIDER_RECORDS_PERIOD.Name) {
		monitorOptions = append(monitorOptions, monitor.WithProviderRecordsPeriod(time.Second*time.Duration(c.Int(FLAG_PROVIDER_RECORDS_PERIOD.Name))))
	}

	if c.IsSet(FLAG_PROVIDER_RECORDS_TIMEOUT.Name) {
		monitorOptions = append(monitorOptions, monitor.WithProviderRecordsTimeout(time.Second*time.Duration(c.Int(FLAG_PROVIDER_RECORDS_TIMEOUT.Name))))
	}
	monitorOptions = append(monitorOptions, monitor.WithExporter(exporter))

	server, err := monitor.NewMonitor(c.Context, monitorOptions...)
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
