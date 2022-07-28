package main

import (
	"github.com/diogo464/ipfs_telemetry/pkg/monitor"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var DiscoverCommand = &cli.Command{
	Name:        "discover",
	Aliases:     []string{},
	Description: "Send a Discover message to a monitor server",
	ArgsUsage:   "PeerID's to send",
	Action:      discoverAction,
}

func discoverAction(c *cli.Context) error {
	pids := make([]peer.ID, 0, c.Args().Len())
	for _, pstr := range c.Args().Slice() {
		p, err := peer.Decode(pstr)
		if err != nil {
			return err
		}
		pids = append(pids, p)
	}

	conn, err := grpc.Dial(c.String(FLAG_ADDRESS.Name), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client := monitor.NewClient(conn)
	for _, p := range pids {
		if err := client.Discover(c.Context, p); err != nil {
			return err
		}
	}

	return nil
}
