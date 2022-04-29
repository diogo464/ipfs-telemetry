package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/diogo464/telemetry/pkg/probe"
	pb "github.com/diogo464/telemetry/pkg/proto/probe"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func main() {
	app := &cli.App{
		Name:     "probe",
		Commands: []*cli.Command{},
		Flags: []cli.Flag{
			FLAG_PROMETHEUS_ADDRESS,
			FLAG_ADDRESS,
			FLAG_NAME,
		},
		Action: mainAction,
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
	defer grpc_server.GracefulStop()

	probeServer, err := probe.NewProbeServer(
		probe.WithName(c.String(FLAG_NAME.Name)),
	)
	if err != nil {
		return err
	}

	pb.RegisterProbeServer(grpc_server, probeServer)

	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logrus.Debug("starting prometheus on ", c.String(FLAG_PROMETHEUS_ADDRESS.Name))
		log.Fatal(http.ListenAndServe(c.String(FLAG_PROMETHEUS_ADDRESS.Name), nil))
	}()

	return probeServer.Run(c.Context)
}
