package main

import (
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
		Name: "telemetry",
		Commands: []*cli.Command{
			CommandSystemInfo,
			CommandUpload,
			CommandDownload,
			CommandWatch,
			CommandTraceRoute,
		},
	}
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func clientFromAddr(addr string) (*telemetry.Client, error) {
	info, err := peer.AddrInfoFromString(addr)
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
}

func printAsJson(v interface{}) {
	marshaled, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(marshaled))
}
