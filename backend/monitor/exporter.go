package monitor

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/monitor"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var _ (monitor.Exporter) = (*natsExporter)(nil)

var (
	Scope = instrumentation.Scope{
		Name:    "telemetry.d464.sh/monitor/nats",
		Version: "0.0.0",
	}

	meter = otel.Meter("telemetry.d464.sh/monitor/nats")

	publishSizeKb, _ = meter.Int64Histogram(
		"publish_size",
		metric.WithDescription("Size of a published export message"),
		metric.WithUnit("kBy"),
	)

	publishSizeLatest, _ = meter.Int64Gauge("publish_size_latest")
)

type natsExporter struct {
	sync.Mutex
	client     *nats.Conn
	logger     *zap.Logger
	inprogress map[peer.ID]*Export
}

func timeNow() time.Time {
	return time.Now().UTC()
}

func defaultExport(p peer.ID) *Export {
	return &Export{
		ObservedAt: timeNow(),
		Peer:       p,
		Session:    [16]byte{},
		Properties: []ExportProperty{},
		Metrics:    []ExportMetrics{},
		Events:     []ExportEvents{},
		Bandwidth:  nil,
	}
}

func newExporter(client *nats.Conn, logger *zap.Logger) *natsExporter {
	return &natsExporter{
		client:     client,
		logger:     logger,
		inprogress: make(map[peer.ID]*Export),
	}
}

// PeerBegin implements monitor.Exporter
func (e *natsExporter) PeerBegin(p peer.ID) {
	e.Lock()
	defer e.Unlock()

	e.inprogress[p] = defaultExport(p)
}

// PeerFailure implements monitor.Exporter
func (e *natsExporter) PeerFailure(p peer.ID, err error) {
	e.Lock()
	defer e.Unlock()

	e.logger.Warn("export failed", zap.String("peer", p.String()), zap.Error(err))
	delete(e.inprogress, p)
}

// PeerSuccess implements monitor.Exporter
func (e *natsExporter) PeerSuccess(p peer.ID) {
	e.Lock()
	defer e.Unlock()

	exp := e.inprogress[p]
	delete(e.inprogress, p)
	if marshaled, err := json.Marshal(exp); err == nil {
		publishSizeKb.Record(context.Background(), int64(len(marshaled))/1024)
		publishSizeLatest.Record(context.Background(), int64(len(marshaled)))
		if err := e.client.Publish(Subject_Export, marshaled); err != nil {
			e.logger.Error("failed to publish telemetry export", zap.Error(err))
		}
	} else {
		e.logger.Error("failed to marshal export", zap.Any("export", exp), zap.Error(err))
	}
}

// Bandwidth implements monitor.Exporter
func (e *natsExporter) Bandwidth(p peer.ID, b telemetry.Bandwidth) {
	e.Lock()
	defer e.Unlock()

	exp := e.getPeerExport(p)
	exp.Bandwidth = &ExportBandwidth{
		UploadRate:   uint64(b.UploadRate),
		DownloadRate: uint64(b.DownloadRate),
	}
}

// Events implements monitor.Exporter
func (e *natsExporter) Events(p peer.ID, s telemetry.Session, d telemetry.EventDescriptor, es []telemetry.Event) {
	e.Lock()
	defer e.Unlock()

	exp := e.getPeerExportWithSess(p, s)
	exp.Events = append(exp.Events, ExportEvents{
		Descriptor: d,
		Events:     es,
	})
}

// Metrics implements monitor.Exporter
func (e *natsExporter) Metrics(p peer.ID, s telemetry.Session, ms telemetry.Metrics) {
	e.Lock()
	defer e.Unlock()

	exp := e.getPeerExportWithSess(p, s)
	for _, m := range ms.OTLP {
		marshaledOtlp, err := proto.Marshal(m)
		if err != nil {
			e.logger.Warn("failed to marshal OTLP metrics", zap.Error(err))
			return
		}
		exp.Metrics = append(exp.Metrics, ExportMetrics{
			OTLP: marshaledOtlp,
		})
	}
}

// Properties implements monitor.Exporter
func (e *natsExporter) Properties(p peer.ID, s telemetry.Session, ps []telemetry.Property) {
	e.Lock()
	defer e.Unlock()

	exp := e.getPeerExportWithSess(p, s)
	for _, p := range ps {
		var value interface{}
		switch v := p.Value.(type) {
		case *telemetry.PropertyValueString:
			vv := v.GetString()
			value = &vv
		case *telemetry.PropertyValueInteger:
			vv := v.GetInteger()
			value = &vv
		default:
			panic("unknown property value type")
		}

		exp.Properties = append(exp.Properties, ExportProperty{
			Scope:       p.Scope,
			Name:        p.Name,
			Description: p.Description,
			Value:       value,
		})
	}
}

// Session implements monitor.Exporter
func (e *natsExporter) Session(p peer.ID, s telemetry.Session) {
	e.Lock()
	defer e.Unlock()

	exp := e.getPeerExport(p)
	exp.Session = s
}

func (e *natsExporter) getPeerExport(p peer.ID) *Export {
	exp := e.inprogress[p]
	if exp == nil {
		e.logger.Fatal("called getPeerExport for peer before PeerBegin", zap.String("peer", p.String()))
	}
	return exp
}

func (e *natsExporter) getPeerExportWithSess(p peer.ID, s telemetry.Session) *Export {
	exp := e.getPeerExport(p)
	exp.Session = s
	return exp
}
