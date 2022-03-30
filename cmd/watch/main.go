package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:        "watch",
		Description: "output snapshot stream from a peer",
		Action:      mainAction,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return err
	}
	fmt.Println("host created")

	info, err := peer.AddrInfoFromString(c.Args().First())
	if err != nil {
		return err
	}
	fmt.Println("added addr info to host")

	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	client := telemetry.NewClient(h, info.ID)
	fmt.Println("client created")

	snapshots, err := client.Snapshots(context.TODO())
	if err != nil {
		return err
	}

	fmt.Println("snapshots retreived")

	for _, snapshot := range snapshots {
		if err := printInterface(snapshot); err != nil {
			return err
		}
	}

	return nil
}

func printInterface(v telemetry.Snapshot) error {
	fmt.Printf("%T\n", v)
	marshaled, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(marshaled))
	return nil
}
