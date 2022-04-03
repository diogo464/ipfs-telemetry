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
var FLAG_STORAGE = &cli.BoolFlag{Name: "storage"}
var FLAG_KADEMLIA = &cli.BoolFlag{Name: "kademlia"}

var CommandWatch = &cli.Command{
	Name:   "watch",
	Action: actionWatch,
	Flags: []cli.Flag{
		FLAG_PING,
		FLAG_RT,
		FLAG_NETWORK,
		FLAG_RESOURCES,
		FLAG_BITSWAP,
		FLAG_STORAGE,
		FLAG_KADEMLIA,
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
	show_storage := c.Bool(FLAG_STORAGE.Name)
	show_kademlia := c.Bool(FLAG_KADEMLIA.Name)
	if !show_ping && !show_rt && !show_network && !show_resources && !show_bitswap && !show_storage && !show_kademlia {
		show_ping = true
		show_rt = true
		show_network = true
		show_resources = true
		show_bitswap = true
		show_storage = true
		show_kademlia = true
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
				case *snapshot.Storage:
					show = show_storage
				case *snapshot.KademliaQuery:
					show = show_kademlia
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
