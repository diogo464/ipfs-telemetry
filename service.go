package telemetry

import (
	"context"
	"fmt"
	"net"

	"github.com/diogo464/telemetry/internal/bpool"
	"github.com/diogo464/telemetry/internal/otlp_exporter"
	"github.com/diogo464/telemetry/internal/pb"
	"github.com/diogo464/telemetry/internal/stream"
	"github.com/diogo464/telemetry/metrics"
	"github.com/libp2p/go-libp2p/core/host"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
)

type ServiceAccessType string

const (
	ServiceAccessPublic     ServiceAccessType = "public"
	ServiceAccessRestricted ServiceAccessType = "restricted"
	ServiceAccessDisabled   ServiceAccessType = "disabled"
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

	serviceAcl *serviceAccessControl
	streams    *serviceStreams
	metrics    *serviceMetrics
	properties *serviceProperties
	events     *serviceEvents

	downloadBlocker *requestBlocker
	uploadBlocker   *requestBlocker

	smetrics *metrics.Metrics
}

func NewService(h host.Host, os ...ServiceOption) (*Service, MeterProvider, error) {
	opts := serviceDefaults()
	err := serviceApply(opts, os...)
	if err != nil {
		return nil, nil, err
	}

	bufferPool := bpool.New(bpool.WithAllocSize(64*1024), bpool.WithMaxSize(8*1024*1024))
	streams := newServiceStreams(
		stream.WithActiveBufferLifetime(opts.activeBufferDuration),
		stream.WithSegmentLifetime(opts.windowDuration),
		stream.WithPool(bufferPool),
	)

	ctx, cancel := context.WithCancel(context.Background())

	t := &Service{
		session:        RandomSession(),
		opts:           opts,
		host:           h,
		meter_provider: nil,
		bufferPool:     bufferPool,

		ctx:    ctx,
		cancel: cancel,

		serviceAcl: nil,
		streams:    streams,
		metrics: newServiceMetrics(streams.create(&pb.StreamType{
			Type: &pb.StreamType_Metric{},
		}).stream),
		properties: newServiceProperties(),
		events:     newServiceEvents(streams),

		downloadBlocker: newRequestBlocker(),
		uploadBlocker:   newRequestBlocker(),
	}

	if opts.enableBandwidth {
		h.SetStreamHandler(ID_UPLOAD, t.uploadHandler)
		h.SetStreamHandler(ID_DOWNLOAD, t.downloadHandler)
	}

	exporter := otlp_exporter.New(t.metrics.stream)
	reader := sdk_metric.NewPeriodicReader(
		exporter,
		sdk_metric.WithInterval(opts.metricsPeriod),
	)

	meter_provider, err := opts.meterProviderFactory(reader)
	if err != nil {
		return nil, nil, err
	}

	t.meter_provider = newServiceMeterProvider(t, meter_provider)

	aclMetrics, err := metrics.NewAclMetrics(t.meter_provider)
	if err != nil {
		return nil, nil, err
	}
	t.serviceAcl = newServiceAccessControl(opts.serviceAccessType, opts.serviceAccessWhitelist, aclMetrics)

	smetrics, err := metrics.NewMetrics(t.meter_provider)
	if err != nil {
		return nil, nil, err
	}
	t.smetrics = smetrics
	t.smetrics.RegisterCallback(func(ctx context.Context) {
		t.smetrics.StreamCount.Observe(ctx, int64(t.streams.getSize()))
		t.smetrics.PropertyCount.Observe(ctx, int64(t.properties.getSize()))
		t.smetrics.EventCount.Observe(ctx, int64(t.events.getSize()))
	})

	streamMetrics, err := metrics.NewStreamMetrics(t.meter_provider)
	if err != nil {
		return nil, nil, err
	}

	streamMetrics.RegisterCallback(func(ctx context.Context) {
		for _, s := range t.streams.getStats() {
			attr := metrics.KeyStreamID.Int(int(s.streamId))
			streamMetrics.UsedSize.Observe(ctx, int64(s.stats.UsedSize), attr)
			streamMetrics.TotalSize.Observe(ctx, int64(s.stats.TotalSize), attr)
		}
	})

	var listener net.Listener
	if opts.listener == nil {
		listener, err = newServiceListener(h, ID_TELEMETRY, t.serviceAcl)
		if err != nil {
			return nil, nil, err
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

	return t, t.meter_provider, nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) Close() {
	s.grpcServer.GracefulStop()
	s.cancel()
}
