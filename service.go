package telemetry

import (
	"context"
	"fmt"
	"net"

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

type Service struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session        Session
	opts           *serviceOptions
	host           host.Host
	meter_provider *serviceMeterProvider
	bufferPool     *bpool.Pool

	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	streams    *serviceStreams
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

	bufferPool := bpool.New(bpool.WithAllocSize(64*1024), bpool.WithMaxSize(8*1024*1024))
	streams := newServiceStreams(
		stream.WithActiveBufferLifetime(opts.activeBufferDuration),
		stream.WithSegmentLifetime(opts.windowDuration),
		stream.WithPool(bufferPool),
	)

	ctx, cancel := context.WithCancel(context.Background())
	session := RandomSession()
	t := &Service{
		session:        session,
		opts:           opts,
		host:           h,
		meter_provider: nil,
		bufferPool:     bufferPool,

		ctx:    ctx,
		cancel: cancel,

		streams:    streams,
		metrics:    newServiceMetrics(streams.create().stream),
		properties: newServiceProperties(),
		captures:   newServiceCaptures(ctx, streams),
		events:     newServiceEvents(streams),

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

	sdk_meter_provider := sdk_metric.NewMeterProvider(
		sdk_metric.WithResource(r),
		sdk_metric.WithReader(
			sdk_metric.NewPeriodicReader(
				otlpmetric.New(t),
				sdk_metric.WithInterval(opts.metricsPeriod),
			),
		),
	)
	t.meter_provider = newServiceMeterProvider(t, sdk_meter_provider)

	return t, nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) Close() {
	s.grpcServer.GracefulStop()
	s.cancel()
}
