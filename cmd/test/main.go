package main

import (
	"context"
	"fmt"
	"net"
	"time"

	pb "git.d464.sh/adc/telemetry/pkg/proto/telemetry"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	h, e := libp2p.New(libp2p.NoListenAddrs)
	die(e)
	p, e := peer.Decode("12D3KooWDtdsFG2TbvK8S2N2xa8c7cbJuVY95zpV1m6TSUrFcEpP")
	die(e)
	m, e := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4001")
	die(e)
	h.Peerstore().AddAddr(p, m, peerstore.PermanentAddrTTL)
	c, e := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithDialer(func(s string, d time.Duration) (net.Conn, error) {
			return gostream.Dial(context.Background(), h, p, telemetry.ID)
		}))
	die(e)
	i := pb.NewClientClient(c)
	info, e := i.GetSystemInfo(context.Background(), &emptypb.Empty{})
	die(e)
	fmt.Println(info)
}

func die(err error) {
	if err != nil {
		panic(err)
	}
}
