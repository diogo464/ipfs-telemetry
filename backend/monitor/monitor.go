package monitor

import (
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

const (
	Subject_Discover = "monitor.discover"
	Subject_Export   = "monitor.export"
)

type DiscoveryMessage struct {
	ID        peer.ID               `json:"id"`
	Addresses []multiaddr.Multiaddr `json:"addresses"`
}

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
