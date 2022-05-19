package telemetry

import (
	"context"
	"fmt"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/telemetry"
	"github.com/diogo464/telemetry/pkg/telemetry/collector"
	"github.com/diogo464/telemetry/pkg/telemetry/config"
	"github.com/diogo464/telemetry/pkg/telemetry/window"
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
	conf config.Config

	ctx              context.Context
	cancel           context.CancelFunc
	grpc_server      *grpc.Server
	boot_time        time.Time
	twindow          window.Window
	twindow_duration time.Duration
	collectors       []collector.Collector

	throttler_upload   *serviceThrottler
	throttler_download *serviceThrottler

	relayCollector *collector.RelayCollector
}

func NewTelemetryService(n *core.IpfsNode, conf config.Config) (*TelemetryService, error) {
	h := n.PeerHost

	defaultConfig := config.Default()
	windowDuration := config.SecondsToDuration(conf.Window.Duration, defaultConfig.Window.Duration)
	windowSize := conf.Window.EventCount
	if windowSize == 0 {
		windowSize = defaultConfig.Window.EventCount
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := RandomSession()
	t := &TelemetryService{
		session:            session,
		node:               n,
		conf:               conf,
		ctx:                ctx,
		cancel:             cancel,
		boot_time:          time.Now().UTC(),
		twindow:            window.NewMemoryWindow(windowDuration, windowSize),
		twindow_duration:   windowDuration,
		collectors:         make([]collector.Collector, 0),
		throttler_upload:   newServiceThrottler(),
		throttler_download: newServiceThrottler(),
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

	err = t.startCollectors()
	if err != nil {
		return nil, err
	}

	t.startEventCollector()

	go metricsTask(t.twindow)

	return t, nil
}

func (s *TelemetryService) Close() {
	s.grpc_server.GracefulStop()
	s.cancel()

	for _, c := range s.collectors {
		c.Close()
	}
}

func (s *TelemetryService) deferCollectorClose(c collector.Collector) {
	s.collectors = append(s.collectors, c)
}

func (s *TelemetryService) startCollectors() error {
	def := config.Default()
	ssink := window.SnapshotSink(s.twindow)

	// ping
	pingCount := s.conf.Ping.Count
	if pingCount == 0 {
		pingCount = 5
	}
	pingCollector := collector.NewPingCollector(s.node.PeerHost, collector.PingOptions{
		PingCount: pingCount,
		Timeout:   config.SecondsToDuration(s.conf.Ping.Timeout, def.Ping.Timeout),
	})
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Ping.Interval, def.Ping.Interval), ssink, pingCollector)
	s.deferCollectorClose(pingCollector)

	// connections
	connectionsCollector := collector.NewConnectionsCollector(s.node.PeerHost)
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Connections.Interval, def.Connections.Interval), ssink, connectionsCollector)
	s.deferCollectorClose(connectionsCollector)

	// network
	networkCollector := collector.NewNetworkCollector(s.node, collector.NetworkOptions{
		BandwidthByPeerInterval: config.SecondsToDuration(s.conf.NetworkCollector.BandwidthByPeerInterval, def.NetworkCollector.BandwidthByPeerInterval),
	})
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.NetworkCollector.Interval, def.NetworkCollector.Interval), ssink, networkCollector)
	s.deferCollectorClose(networkCollector)

	// routing table
	routingTableCollector := collector.NewRoutingTableCollector(s.node)
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.RoutingTable.Interval, def.RoutingTable.Interval), ssink, routingTableCollector)
	s.deferCollectorClose(routingTableCollector)

	// resources
	resourcesCollector, err := collector.NewResourcesCollector()
	if err != nil {
		return err
	}
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Resources.Interval, def.Resources.Interval), ssink, resourcesCollector)
	s.deferCollectorClose(resourcesCollector)

	// bitswap
	bitswapCollector := collector.NewBitswapCollector(s.node)
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Bitswap.Interval, def.Bitswap.Interval), ssink, bitswapCollector)
	s.deferCollectorClose(bitswapCollector)

	// storage
	storageCollector := collector.NewStorageCollector(s.node)
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Storage.Interval, def.Storage.Interval), ssink, storageCollector)
	s.deferCollectorClose(storageCollector)

	// kademlia
	kademliaCollector := collector.NewKademliaCollector()
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Kademlia.Interval, def.Kademlia.Interval), ssink, kademliaCollector)
	s.deferCollectorClose(kademliaCollector)

	// traceroute
	tracerouteCollector := collector.NewTracerouteCollector(s.node.PeerHost)
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.TraceRoute.Interval, def.TraceRoute.Interval), ssink, tracerouteCollector)
	s.deferCollectorClose(tracerouteCollector)

	// window
	windowCollector := collector.NewWindowCollector(s.twindow_duration, s.twindow)
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Window.Interval, def.Window.Interval), ssink, windowCollector)
	s.deferCollectorClose(windowCollector)

	// relay
	relayCollector := collector.NewRelayCollector()
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.Relay.Interval, def.Relay.Interval), ssink, relayCollector)
	s.deferCollectorClose(relayCollector)
	s.relayCollector = relayCollector

	// holepunch
	holePunchCollector := collector.NewHolePunchCollector()
	collector.RunCollector(s.ctx, config.SecondsToDuration(s.conf.HolePunch.Interval, def.HolePunch.Interval), ssink, holePunchCollector)
	s.deferCollectorClose(holePunchCollector)

	return nil
}

func (s *TelemetryService) startEventCollector() {
	esink := window.EventSink(s.twindow)

	collector.StartKademliaEventCollector(s.ctx, esink)

	s.relayCollector.SetEventSink(esink)
}
