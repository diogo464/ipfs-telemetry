package main

import (
	"context"
	"fmt"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
)

var LOCAL_PEER peer.ID

func init() {
	pid, err := peer.Decode("12D3KooWDmpSWyXdjFYhdSfvNiz48adEABZfrLDcfguTsWyriRM9")
	die(err)
	LOCAL_PEER = pid
}

func main() {
	h := createHost()
	client, err := telemetry.NewClient(h, LOCAL_PEER)
	die(err)

	for {
		info, err := client.SystemInfo(context.Background())
		if err == nil {
			fmt.Println(info)
		} else {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 3)
	}
}

func createHost() host.Host {
	info, err := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWDmpSWyXdjFYhdSfvNiz48adEABZfrLDcfguTsWyriRM9")
	die(err)

	h, err := libp2p.New(libp2p.NoListenAddrs)
	die(err)
	h.Peerstore().SetAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	return h
}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
