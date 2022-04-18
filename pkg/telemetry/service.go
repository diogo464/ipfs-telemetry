package telemetry

import (
	"context"
	"fmt"
	"time"

	"git.d464.sh/adc/telemetry/pkg/collector"
	pb "git.d464.sh/adc/telemetry/pkg/proto/telemetry"
	"git.d464.sh/adc/telemetry/pkg/telemetry/window"
	"github.com/ipfs/go-ipfs/core"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
)

type TelemetryService struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session Session
	// the node we are collecting telemetry from
	node *core.IpfsNode
	// read-only options
	opts        *options
	cancel      context.CancelFunc
	grpc_server *grpc.Server
	boot_time   time.Time

	snapshots window.Window
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
	session := RandomSession()
	t := &TelemetryService{
		session:   session,
		node:      n,
		opts:      o,
		cancel:    cancel,
		boot_time: time.Now().UTC(),
		snapshots: window.NewMemoryWindow(o.windowDuration),
	}
	h.SetStreamHandler(ID_UPLOAD, t.uploadHandler)
	h.SetStreamHandler(ID_DOWNLOAD, t.downloadHandler)

	listener, err := gostream.Listen(h, ID_TELEMETRY)
	if err != nil {
		return nil, err
	}

	grpc_server := grpc.NewServer()
	pb.RegisterTelemetryServer(grpc_server, t)
	t.grpc_server = grpc_server

	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("GRPC server stopped")
	}()

	go collector.RunPingCollector(ctx, t.node.PeerHost, t.snapshots, collector.PingOptions{
		PingCount: 5,
		Interval:  time.Second * 5,
		Timeout:   time.Second * 10,
	})

	go collector.RunNetworkCollector(ctx, t.node, t.snapshots, collector.NetworkOptions{
		Interval:                time.Second * 30,
		BandwidthByPeerInterval: time.Minute * 5,
	})

	go collector.RunRoutingTableCollector(ctx, t.node, t.snapshots, collector.RoutingTableOptions{
		Interval: time.Second * 60,
	})

	go collector.RunResourcesCollector(ctx, t.snapshots, collector.ResourcesOptions{
		Interval: time.Second * 10,
	})

	go collector.RunBitswapCollector(ctx, t.node, t.snapshots, collector.BitswapOptions{
		Interval: time.Second * 30,
	})

	go collector.RunStorageCollector(ctx, t.node, t.snapshots, collector.StorageOptions{
		Interval: time.Second * 60,
	})

	go collector.RunKademliaCollector(ctx, t.snapshots, collector.KademliaOptions{
		Interval: time.Second * 30,
	})

	go collector.RunTraceRouteCollector(ctx, h, t.snapshots, collector.TraceRouteOptions{
		Interval: time.Second * 5,
	})

	go collector.RunWindowCollector(ctx, o.windowDuration, t.snapshots, t.snapshots, collector.StorageOptions{
		Interval: time.Second * 5,
	})

	go metricsTask(t.snapshots)

	return t, nil
}

func (s *TelemetryService) Close() {
	s.grpc_server.GracefulStop()
	s.cancel()
}
