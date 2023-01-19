package telemetry

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/diogo464/telemetry/internal/bpool"
	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
)

var (
	_ (otlpmetric.Client) = (*Service)(nil)
)

type serviceMetrics struct {
	mu          sync.Mutex
	stream      *stream.Stream
	descriptors []*pb.MetricDescriptor
}

type servicePropertyEntry struct {
	pbproperty *pb.Property
}

type serviceProperties struct {
	mu         sync.Mutex
	properties []*servicePropertyEntry
}

type serviceCaptures struct {
	mu       sync.Mutex
	captures map[uint32]*serviceCapture
	nextID   uint32
}

type serviceEvent struct {
	config     eventConfig
	stream     *stream.Stream
	emitter    *eventEmitter
	descriptor *pb.EventDescriptor
}

type serviceEvents struct {
	mu     sync.Mutex
	events map[uint32]*serviceEvent
	nextID uint32
}

type Service struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session        Session
	opts           *serviceOptions
	host           host.Host
	meter_provider *sdk_metric.MeterProvider
	bufferPool     *bpool.Pool

	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	metrics    *serviceMetrics
	properties *serviceProperties
	captures   *serviceCaptures
	events     *serviceEvents

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
		session:        session,
		opts:           opts,
		host:           h,
		meter_provider: nil,
		bufferPool:     bpool.New(bpool.WithAllocSize(64*1024), bpool.WithMaxSize(8*1024*1024)),

		ctx:    ctx,
		cancel: cancel,

		downloadBlocker: newRequestBlocker(),
		uploadBlocker:   newRequestBlocker(),
	}

	t.metrics = &serviceMetrics{
		stream:      t.newStream(),
		descriptors: make([]*pb.MetricDescriptor, 0),
	}
	t.properties = &serviceProperties{
		properties: make([]*servicePropertyEntry, 0),
	}
	t.captures = &serviceCaptures{
		captures: make(map[uint32]*serviceCapture),
	}
	t.events = &serviceEvents{
		events: make(map[uint32]*serviceEvent, 0),
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
		log.Info("grpc server stopped")
	}()

	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceVersionKey.String("v0.0.0"),
			attribute.String("peerid", h.ID().String()),
		),
	)

	t.meter_provider = sdk_metric.NewMeterProvider(
		sdk_metric.WithResource(r),
		sdk_metric.WithReader(
			sdk_metric.NewPeriodicReader(
				otlpmetric.New(t),
				sdk_metric.WithInterval(opts.metricsPeriod),
			),
		),
	)

	return t, nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) Close() {
	s.grpcServer.GracefulStop()
	s.cancel()
}

func (s *Service) newStream() *stream.Stream {
	return stream.New(
		stream.WithActiveBufferLifetime(s.opts.activeBufferDuration),
		stream.WithSegmentLifetime(s.opts.windowDuration),
		stream.WithPool(s.bufferPool),
	)
}
