package telemetry

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/diogo464/telemetry/pb"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
)

var (
	_ (Telemetry)         = (*Service)(nil)
	_ (otlpmetric.Client) = (*Service)(nil)
)

type serviceMetrics struct {
	stream *Stream
}

type servicePropertyEntry struct {
	pbproperty *pb.Property
}

type serviceProperties struct {
	mu         sync.Mutex
	properties map[string]*servicePropertyEntry
}

type serviceCaptures struct {
	mu       sync.Mutex
	captures map[string]*serviceCapture
}

type serviceEvent struct {
	config     EventConfig
	stream     *Stream
	emitter    *eventEmitter
	descriptor *pb.EventDescriptor
}

type serviceEvents struct {
	mu     sync.Mutex
	events map[string]*serviceEvent
}

type Service struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session        Session
	opts           *serviceOptions
	host           host.Host
	meter_provider *sdk_metric.MeterProvider

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

		ctx:    ctx,
		cancel: cancel,

		metrics: &serviceMetrics{
			stream: NewStream(opts.defaultStreamOptions...),
		},
		properties: &serviceProperties{
			properties: make(map[string]*servicePropertyEntry),
		},
		captures: &serviceCaptures{
			captures: make(map[string]*serviceCapture),
		},
		events: &serviceEvents{
			events: make(map[string]*serviceEvent, 0),
		},

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

	t.meter_provider = sdk_metric.NewMeterProvider(sdk_metric.WithReader(sdk_metric.NewPeriodicReader(otlpmetric.New(t), sdk_metric.WithInterval(opts.metricsPeriod))))

	return t, nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) Close() {
	s.grpcServer.GracefulStop()
	s.cancel()
}
