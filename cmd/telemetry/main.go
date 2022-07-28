package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "telemetry",
		Flags: []cli.Flag{
			FLAG_CONN_TYPE,
			FLAG_HOST,
		},
		Commands: []*cli.Command{
			CommandUpload,
			CommandDownload,
			CommandDebug,
			CommandSessionInfo,
			CommandSystemInfo,
			CommandStream,
			CommandStreams,
			CommandProperty,
			CommandProperties,
			CommandWalk,
			//CommandWatch,
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func clientFromContext(c *cli.Context) (*telemetry.Client, error) {
	switch c.String(FLAG_CONN_TYPE.Name) {
	case "libp2p":

		info, err := peer.AddrInfoFromString(c.String(FLAG_HOST.Name))
		if err != nil {
			return nil, err
		}
		h, err := libp2p.New(libp2p.NoListenAddrs)
		if err != nil {
			return nil, err
		}
		fmt.Println(info)
		h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
		return telemetry.NewClient(h, info.ID)
	case "tcp":
		return telemetry.NewClient2(c.String(FLAG_HOST.Name))
	default:
		return nil, fmt.Errorf("Unknown connection type: %s", c.String(FLAG_CONN_TYPE.Name))
	}
}

func printAsJson(v interface{}) {
	marshaled, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshaled))
}
