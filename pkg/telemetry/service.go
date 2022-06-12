package telemetry

import (
	"context"
	"fmt"
	"time"

	pb "github.com/diogo464/telemetry/pkg/proto/telemetry"
	"github.com/ipfs/go-ipfs/core"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"google.golang.org/grpc"
)

type serviceStreamEntry struct {
	blocker *requestBlocker
	stream  *Stream
}

type Service struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session Session
	// the node we are collecting telemetry from
	node *core.IpfsNode
	opts *serviceOptions

	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server
	bootTime   time.Time
	streams    map[string]*serviceStreamEntry

	provRecordsBlocker *requestBlocker
	downloadBlocker    *requestBlocker
	uploadBlocker      *requestBlocker
}

func NewService(n *core.IpfsNode, os ...ServiceOption) (*Service, error) {
	h := n.PeerHost
	opts := serviceDefaults()
	err := serviceApply(opts, os...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := RandomSession()
	t := &Service{
		session:  session,
		node:     n,
		opts:     opts,
		ctx:      ctx,
		cancel:   cancel,
		bootTime: time.Now().UTC(),
		streams:  make(map[string]*serviceStreamEntry),

		provRecordsBlocker: newRequestBlocker(),
		downloadBlocker:    newRequestBlocker(),
		uploadBlocker:      newRequestBlocker(),
	}

	if opts.enableBandwidth {
		h.SetStreamHandler(ID_UPLOAD, t.uploadHandler)
		h.SetStreamHandler(ID_DOWNLOAD, t.downloadHandler)
	}

	listener, err := gostream.Listen(h, ID_TELEMETRY)
	if err != nil {
		return nil, err
	}

	grpc_server := grpc.NewServer()
	pb.RegisterTelemetryServer(grpc_server, t)
	t.grpcServer = grpc_server

	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("GRPC server stopped")
	}()

	return t, nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) RegisterCollector(streamName string, collector Collector, opts CollectorOpts) error {
	if _, ok := s.streams[streamName]; ok {
		return ErrCollectorAlreadyRegistered
	}

	stream := NewStream(s.opts.defaultStreamOptions...)
	s.streams[streamName] = &serviceStreamEntry{
		blocker: newRequestBlocker(),
		stream:  stream,
	}
	go collectorMainLoop(s.ctx, stream, collector, opts)

	return nil
}

func (s *Service) Close() {
	s.grpcServer.GracefulStop()
	s.cancel()
}
