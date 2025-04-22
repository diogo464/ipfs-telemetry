package main

import (
	"encoding/json"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/diogo464/telemetry/monitor"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var _ (monitor.Exporter) = (*natsExporter)(nil)

const (
	TELEMETRY_EXPORT_SUBJECT = "telemetry.export"
	KIND_SESSION             = "session"
	KIND_METRICS             = "metrics"
	KIND_PROPERTIES          = "properties"
	KIND_EVENTS              = "events"
	KIND_BANDWIDTH           = "bandwidth"
)

type ExportBandwidth struct {
	UploadRate   uint64 `json:"upload_rate"`
	DownloadRate uint64 `json:"download_rate"`
}

type ExportEvents struct {
	Descriptor telemetry.EventDescriptor `json:"descriptor"`
	Events     []telemetry.Event         `json:"events"`
}

type ExportMetrics struct {
	OTLP []byte `json:"otlp"`
}

type ExportProperty struct {
	Scope       instrumentation.Scope `json:"scope"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Value       interface{}           `json:"value"`
}

type Export struct {
	ObservedAt time.Time         `json:"observed_at"`
	Peer       peer.ID           `json:"peer"`
	Session    telemetry.Session `json:"session"`

	// Export Data
	Properties []ExportProperty `json:"properties"`
	Metrics    []ExportMetrics  `json:"metrics"`
	Events     []ExportEvents   `json:"events"`
	Bandwidth  *ExportBandwidth `json:"bandwidth"`
}

type natsExporter struct {
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

func newNatsExporter(client *nats.Conn, logger *zap.Logger) *natsExporter {
	return &natsExporter{
		client:     client,
		logger:     logger,
		inprogress: make(map[peer.ID]*Export),
	}
}

// PeerBegin implements monitor.Exporter
func (e *natsExporter) PeerBegin(p peer.ID) {
	e.inprogress[p] = defaultExport(p)
}

// PeerFailure implements monitor.Exporter
func (e *natsExporter) PeerFailure(p peer.ID, err error) {
	e.logger.Warn("export failed", zap.String("peer", p.String()), zap.Error(err))
	delete(e.inprogress, p)
}

// PeerSuccess implements monitor.Exporter
func (e *natsExporter) PeerSuccess(p peer.ID) {
	exp := e.inprogress[p]
	delete(e.inprogress, p)
	if marshaled, err := json.Marshal(exp); err == nil {
		if err := e.client.Publish(TELEMETRY_EXPORT_SUBJECT, marshaled); err != nil {
			e.logger.Error("failed to publish telemetry export", zap.Error(err))
		}
	} else {
		e.logger.Error("failed to marshal export", zap.Any("export", exp), zap.Error(err))
	}
}

// Bandwidth implements monitor.Exporter
func (e *natsExporter) Bandwidth(p peer.ID, b telemetry.Bandwidth) {
	exp := e.getPeerExport(p)
	exp.Bandwidth = &ExportBandwidth{
		UploadRate:   uint64(b.UploadRate),
		DownloadRate: uint64(b.DownloadRate),
	}
}

// Events implements monitor.Exporter
func (e *natsExporter) Events(p peer.ID, s telemetry.Session, d telemetry.EventDescriptor, es []telemetry.Event) {
	exp := e.getPeerExportWithSess(p, s)
	exp.Events = append(exp.Events, ExportEvents{
		Descriptor: d,
		Events:     es,
	})
}

// Metrics implements monitor.Exporter
func (e *natsExporter) Metrics(p peer.ID, s telemetry.Session, ms telemetry.Metrics) {
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
