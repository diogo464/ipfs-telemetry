package telemetry

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/diogo464/telemetry/pb"
	gostream "github.com/libp2p/go-libp2p-gostream"
	"github.com/libp2p/go-libp2p/core/host"
	"google.golang.org/grpc"
)

type serviceMetrics struct {
	collectors map[MetricCollector]struct{}
	stream     *Stream
	// Context to stop the goroutine that is periodically collecting metrics
	ctx context.Context
}

type serviceEventsEntry struct {
	collector EventCollector
	stream    *Stream
}

type serviceEvents struct {
	events map[string]*serviceEventsEntry
}

type serviceSnapshotEntry struct {
	collector  SnapshotCollector
	descriptor SnapshotDescriptor
	stream     *Stream
	// Context to stop the goroutine that is taking the period snapshots
	ctx context.Context
}

type serviceSnapshots struct {
	snapshots map[string]*serviceSnapshotEntry
}

type servicePropertyEntry struct {
	collector  PropertyCollector
	descriptor PropertyDescriptor
}

type serviceProperties struct {
	properties map[string]*servicePropertyEntry
}

type Service struct {
	pb.UnimplementedTelemetryServer
	// current session, randomly generated uuid
	session Session
	opts    *serviceOptions
	h       host.Host
	running bool

	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	metrics    *serviceMetrics
	events     *serviceEvents
	snapshots  *serviceSnapshots
	properties *serviceProperties

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
		running: false,

		metrics: &serviceMetrics{
			collectors: make(map[MetricCollector]struct{}),
			stream:     NewStream(opts.defaultStreamOptions...),
			ctx:        ctx,
		},
		events: &serviceEvents{
			events: make(map[string]*serviceEventsEntry),
		},
		snapshots: &serviceSnapshots{
			snapshots: make(map[string]*serviceSnapshotEntry),
		},
		properties: &serviceProperties{
			properties: make(map[string]*servicePropertyEntry),
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

	return t, nil
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) RegisterMetric(metric MetricCollector) error {
	if _, ok := s.metrics.collectors[metric]; ok {
		return ErrMetricAlreadyRegistered
	}
	s.metrics.collectors[metric] = struct{}{}
	return nil
}

func (s *Service) RegisterMetrics(metrics []MetricCollector) error {
	for _, metric := range metrics {
		if err := s.RegisterMetric(metric); err != nil {
			return nil
		}
	}
	return nil
}

func (s *Service) RegisterEvent(event EventCollector) error {
	descriptor := event.Descriptor()
	if _, ok := s.events.events[descriptor.Name]; ok {
		return ErrEventAlreadyRegistered
	}
	s.events.events[descriptor.Name] = &serviceEventsEntry{
		collector: event,
		stream:    NewStream(s.opts.defaultStreamOptions...),
	}
	return nil
}

func (s *Service) RegisterEvents(events []EventCollector) error {
	for _, event := range events {
		if err := s.RegisterEvent(event); err != nil {
			return nil
		}
	}
	return nil
}

func (s *Service) RegisterSnapshot(snapshot SnapshotCollector) error {
	descriptor := snapshot.Descriptor()
	if _, ok := s.snapshots.snapshots[descriptor.Name]; ok {
		return ErrSnapshotAlreadyRegistered
	}
	s.snapshots.snapshots[descriptor.Name] = &serviceSnapshotEntry{
		collector:  snapshot,
		descriptor: descriptor,
		stream:     NewStream(s.opts.defaultStreamOptions...),
		ctx:        s.ctx,
	}
	return nil
}

func (s *Service) RegisterSnapshots(snapshots []SnapshotCollector) error {
	for _, snapshot := range snapshots {
		if err := s.RegisterSnapshot(snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) RegisterProperty(property PropertyCollector) error {
	descriptor := property.Descriptor()
	if _, ok := s.properties.properties[descriptor.Name]; ok {
		return ErrSnapshotAlreadyRegistered
	}
	s.properties.properties[descriptor.Name] = &servicePropertyEntry{
		collector:  property,
		descriptor: descriptor,
	}
	return nil
}

func (s *Service) RegisterProperties(properties []PropertyCollector) error {
	for _, property := range properties {
		if err := s.RegisterProperty(property); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Start() {
	if s.running {
		return
	}
	s.running = true

	go s.metricsCollectionTask()
	for _, event := range s.events.events {
		event.collector.Open(event.stream)
	}
	for snapshotName := range s.snapshots.snapshots {
		go s.snapshotCollectionTask(snapshotName)
	}
}

func (s *Service) metricsCollectionTask() {
	ticker := time.NewTicker(s.opts.metricsPeriod)
	metrics := s.metrics

LOOP:
	for {
		select {
		case <-metrics.ctx.Done():
			break LOOP
		case <-ticker.C:
		}

		ch := make(chan CollectedMetric, 8)
		m := make(map[string]float64)
		go func() {
			for c := range metrics.collectors {
				c.Collect(ch)
			}
			close(ch)
		}()
		for cm := range ch {
			m[cm.Name] = cm.Value
		}
		mpb := pb.MetricsObservations{
			Observations: m,
		}
		size := mpb.Size()
		metrics.stream.AllocAndWrite(size, func(b []byte) error {
			_, err := mpb.MarshalToSizedBuffer(b)
			return err
		})
	}
}

func (s *Service) snapshotCollectionTask(name string) {
	snapshot := s.snapshots.snapshots[name]
	ticker := time.NewTicker(snapshot.descriptor.Period)

LOOP:
	for {
		select {
		case <-snapshot.ctx.Done():
			break LOOP
		case <-ticker.C:
		}

		if err := snapshot.collector.Collect(snapshot.stream); err != nil {
			log.Warnf("snapshot collection error",
				"snapshot", snapshot.descriptor.Name,
				"error", err)
		}
	}
}

func (s *Service) Close() {
	s.running = false

	for _, event := range s.events.events {
		event.collector.Close()
	}

	s.grpcServer.GracefulStop()
	s.cancel()
}
