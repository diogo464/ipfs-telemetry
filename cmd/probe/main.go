package main

import (
	"fmt"
	"net"
	"os"

	"git.d464.sh/adc/telemetry/pkg/probe"
	pb "git.d464.sh/adc/telemetry/pkg/proto/probe"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func main() {
	app := &cli.App{
		Name:     "probe",
		Commands: []*cli.Command{},
		Flags: []cli.Flag{
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

	return probeServer.Run(c.Context)
}
