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
	"go.opentelemetry.io/contrib/bridges/prometheus"
	"go.opentelemetry.io/otel/metric"
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
		metrics:    newServiceMetrics(streams),
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
		sdk_metric.WithProducer(prometheus.NewMetricProducer()),
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
	t.smetrics.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(t.smetrics.StreamCount, int64(t.streams.getSize()))
		obs.ObserveInt64(t.smetrics.PropertyCount, int64(t.properties.getSize()))
		obs.ObserveInt64(t.smetrics.EventCount, int64(t.events.getSize()))
		return nil
	})

	streamMetrics, err := metrics.NewStreamMetrics(t.meter_provider)
	if err != nil {
		return nil, nil, err
	}

	streamMetrics.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		for _, s := range t.streams.getStats() {
			attr := metrics.KeyStreamID.Int(int(s.streamId))
			obs.ObserveInt64(streamMetrics.UsedSize, int64(s.stats.UsedSize), metric.WithAttributes(attr))
			obs.ObserveInt64(streamMetrics.TotalSize, int64(s.stats.TotalSize), metric.WithAttributes(attr))
		}
		return nil
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
