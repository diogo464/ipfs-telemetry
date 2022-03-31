package telemetry

import (
	"fmt"
	"time"

	"git.d464.sh/adc/telemetry/plugin/pb"
	"git.d464.sh/adc/telemetry/plugin/snapshot"
	"git.d464.sh/adc/telemetry/plugin/window"
	"github.com/google/uuid"
	"github.com/ipfs/go-ipfs/core"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
)

const (
	ID                     = "/telemetry/telemetry/0.0.0"
	BANDWIDTH_PAYLOAD_SIZE = 32 * 1024 * 1024
)

type TelemetryService struct {
	pb.UnimplementedClientServer
	// current session, randomly generated number
	s uuid.UUID
	// the node we are collecting telemetry from
	n *core.IpfsNode
	// read-only options
	o *options
	w window.Window
}

func NewTelemetryService(n *core.IpfsNode, opts ...Option) (*TelemetryService, error) {
	o := new(options)
	defaults(o)

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	t := &TelemetryService{
		s: uuid.New(),
		n: n,
		o: o,
		w: window.NewWindow(o.windowDuration),
	}

	h := n.PeerHost
	listener, err := gostream.Listen(h, ID)
	if err != nil {
		return nil, err
	}

	grpc_server := grpc.NewServer()
	pb.RegisterClientServer(grpc_server, t)

	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go snapshot.NewPingCollector(t.n.PeerHost, t.w, snapshot.PingOptions{
		PingCount: 5,
		Interval:  time.Second * 5,
		Timeout:   time.Second * 10,
	}).Run()

	go snapshot.NewRoutingTableCollector(t.n, t.w, snapshot.RoutingTableOptions{
		Interval: time.Second * 3,
	}).Run()

	go snapshot.NewNetworkCollector(t.n, t.w, snapshot.NetworkOptions{
		Interval: time.Second * 15,
	}).Run()

	return t, nil
}
