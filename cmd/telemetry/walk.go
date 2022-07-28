package main

import (
	"fmt"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/walker"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var FLAG_WALK_PROTOCOL = &cli.StringFlag{
	Name:  "protocol",
	Value: telemetry.ID_TELEMETRY,
	Usage: "The protocol to search for",
}

var CommandWalk = &cli.Command{
	Name:        "walk",
	Description: "walk dht to find peer supporting the telemetry protocol",
	Action:      actionWalk,
	Flags:       []cli.Flag{FLAG_WALK_PROTOCOL},
}

func actionWalk(c *cli.Context) error {
	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return err
	}

	targetProto := c.String(FLAG_WALK_PROTOCOL.Name)
	w, err := walker.New(
		h,
		walker.WithObserver(walker.NewPeerObserverFn(func(p *walker.Peer) {
			for _, proto := range p.Protocols {
				if proto == targetProto {
					for _, addr := range p.Addresses {
						p2p, err := multiaddr.NewComponent("p2p", p.ID.Pretty())
						if err != nil {
							addr = addr.Encapsulate(p2p)
							fmt.Println(addr.String())
						}
					}
					break
				}
			}
		})),
	)
	if err != nil {
		return err
	}

	return w.Walk(c.Context)
}
