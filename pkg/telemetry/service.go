package telemetry

import (
	"context"
	"fmt"
	"time"

	"git.d464.sh/adc/telemetry/pkg/collector"
	pb "git.d464.sh/adc/telemetry/pkg/proto/telemetry"
	"git.d464.sh/adc/telemetry/pkg/window"
	"github.com/google/uuid"
	"github.com/ipfs/go-ipfs/core"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
)

type TelemetryService struct {
	pb.UnimplementedClientServer
	// current session, randomly generated number
	session uuid.UUID
	// the node we are collecting telemetry from
	node *core.IpfsNode
	// read-only options
	opts        *options
	cancel      context.CancelFunc
	wnd         window.Window
	grpc_server *grpc.Server
}

func NewTelemetryService(n *core.IpfsNode, opts ...Option) (*TelemetryService, error) {
	o := new(options)
	defaults(o)

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	h := n.PeerHost

	ctx, cancel := context.WithCancel(context.Background())
	t := &TelemetryService{
		session: uuid.New(),
		node:    n,
		opts:    o,
		cancel:  cancel,
		wnd:     window.NewWindow(o.windowDuration),
	}
	h.SetStreamHandler(ID_UPLOAD, t.uploadHandler)
	h.SetStreamHandler(ID_DOWNLOAD, t.downloadHandler)

	listener, err := gostream.Listen(h, ID_TELEMETRY)
	if err != nil {
		return nil, err
	}

	grpc_server := grpc.NewServer()
	pb.RegisterClientServer(grpc_server, t)
	t.grpc_server = grpc_server

	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("GRPC server stopped")
	}()

	go collector.RunPingCollector(ctx, t.node.PeerHost, t.wnd, collector.PingOptions{
		PingCount: 5,
		Interval:  time.Second * 5,
		Timeout:   time.Second * 10,
	})

	go collector.RunNetworkCollector(ctx, t.node, t.wnd, collector.NetworkOptions{
		Interval: time.Second * 3,
	})

	go collector.RunNetworkCollector(ctx, t.node, t.wnd, collector.NetworkOptions{
		Interval: time.Second * 15,
	})

	go collector.RunRoutintTableCollector(ctx, t.node, t.wnd, collector.RoutingTableOptions{
		Interval: time.Second * 10,
	})

	go collector.RunResourcesCollector(ctx, t.wnd, collector.ResourcesOptions{
		Interval: time.Second * 2,
	})

	go collector.RunBitswapCollector(ctx, t.node, t.wnd, collector.BitswapOptions{
		Interval: time.Second * 5,
	})

    go collector.RunStorageCollector(ctx, t.node, t.wnd, collector.StorageOptions{
        Interval: time.Second * 5,
    })

	return t, nil
}

func (s *TelemetryService) Close() {
	s.grpc_server.GracefulStop()
	s.cancel()
}
