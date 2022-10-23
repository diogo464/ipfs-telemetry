package telemetry

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/diogo464/telemetry/pb"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"google.golang.org/grpc"
)

type serviceStreamEntry struct {
	blocker      *requestBlocker
	stream       *Stream
	descriptor   *CollectorDescriptor
	observers_mu sync.Mutex
	observers    map[chan<- struct{}]struct{}
}

type servicePropertyEntry struct {
	property   Property
	descriptor *PropertyDescriptor
}

type Service struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session Session
	opts    *serviceOptions
	h       host.Host

	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	// all modifications should be done at the start so that no locking is required
	streams    map[string]*serviceStreamEntry
	properties map[string]*servicePropertyEntry

	downloadBlocker *requestBlocker
	uploadBlocker   *requestBlocker
}

func NewService(h host.Host, os ...ServiceOption) (*Service, error) {
	opts := serviceDefaults()
	err := serviceApply(opts, os...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	session := RandomSession()
	t := &Service{
		session: session,
		opts:    opts,
		h:       h,
		ctx:     ctx,
		cancel:  cancel,

		streams:    make(map[string]*serviceStreamEntry),
		properties: make(map[string]*servicePropertyEntry),

		downloadBlocker: newRequestBlocker(),
		uploadBlocker:   newRequestBlocker(),
	}

	if opts.enableBandwidth {
		h.SetStreamHandler(ID_UPLOAD, t.uploadHandler)
		h.SetStreamHandler(ID_DOWNLOAD, t.downloadHandler)
	}

	var listener net.Listener
	if opts.listener == nil {
		listener, err = gostream.Listen(h, ID_TELEMETRY)
		if err != nil {
			return nil, err
		}
	} else {
		listener = opts.listener
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

func (s *Service) Register(collector Collector, opts ...CollectorOption) error {
	config := collectorConfigDefaults()
	if err := collectorConfigApply(config, opts...); err != nil {
		return err
	}

	descriptor := collector.Descriptor()
	if _, ok := s.streams[descriptor.Name]; ok {
		return ErrCollectorAlreadyRegistered
	}

	if config.name != nil {
		descriptor.Name = *config.name
	}
	if config.period != nil {
		descriptor.Period = *config.period
	}
	if config.encoding != nil {
		descriptor.Encoding = *config.encoding
	}

	stream := NewStream(s.opts.defaultStreamOptions...)
	s.streams[descriptor.Name] = &serviceStreamEntry{
		blocker:    newRequestBlocker(),
		stream:     stream,
		descriptor: &descriptor,
		observers:  make(map[chan<- struct{}]struct{}),
	}
	go s.collectorMainLoop(s.ctx, stream, collector, descriptor)

	return nil
}

func (s *Service) RegisterProperty(property Property, opts ...PropertyOption) error {
	config := propertyConfigDefaults()
	if err := propertyConfigApply(config, opts...); err != nil {
		return err
	}

	descriptor := property.Descriptor()
	if _, ok := s.properties[descriptor.Name]; ok {
		return ErrPropertyAlreadyRegistered
	}

	if config.overrideName != nil {
		descriptor.Name = *config.overrideName
	}
	if config.overrideEncoding != nil {
		descriptor.Encoding = *config.overrideEncoding
	}

	s.properties[descriptor.Name] = &servicePropertyEntry{
		property:   property,
		descriptor: &descriptor,
	}

	return nil
}

func (s *Service) Close() {
	s.grpcServer.GracefulStop()
	s.cancel()
}
