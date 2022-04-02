package main

import (
	"time"

	"git.d464.sh/adc/telemetry/pkg/snapshot"
	"github.com/urfave/cli/v2"
)

var FLAG_PING = &cli.BoolFlag{Name: "ping"}
var FLAG_RT = &cli.BoolFlag{Name: "rt"}
var FLAG_NETWORK = &cli.BoolFlag{Name: "network"}
var FLAG_RESOURCES = &cli.BoolFlag{Name: "resources"}
var FLAG_BITSWAP = &cli.BoolFlag{Name: "bitswap"}

var CommandWatch = &cli.Command{
	Name:   "watch",
	Action: actionWatch,
	Flags: []cli.Flag{
		FLAG_PING,
		FLAG_RT,
		FLAG_NETWORK,
		FLAG_RESOURCES,
		FLAG_BITSWAP,
	},
}

func actionWatch(c *cli.Context) error {
	client, err := clientFromAddr(c.Args().First())
	if err != nil {
		return err
	}
	defer client.Close()

	show_ping := c.Bool(FLAG_PING.Name)
	show_rt := c.Bool(FLAG_RT.Name)
	show_network := c.Bool(FLAG_NETWORK.Name)
	show_resources := c.Bool(FLAG_RESOURCES.Name)
	show_bitswap := c.Bool(FLAG_BITSWAP.Name)
	if !show_ping && !show_rt && !show_network && !show_resources && !show_bitswap {
		show_ping = true
		show_rt = true
		show_network = true
		show_resources = true
		show_bitswap = true
	}

	ticker := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-ticker.C:
			snapshots, err := client.Snapshots(c.Context)
			if err != nil {
				return err
			}

			for _, s := range snapshots {
				var show bool = false
				switch s.(type) {
				case *snapshot.Ping:
					show = show_ping
				case *snapshot.RoutingTable:
					show = show_rt
				case *snapshot.Network:
					show = show_network
				case *snapshot.Resources:
					show = show_resources
				case *snapshot.Bitswap:
					show = show_bitswap
				}
				if show {
					printAsJson(s)
				}
			}

		case <-c.Context.Done():
			break LOOP
		}
	}

	return nil
}
