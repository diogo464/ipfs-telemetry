package config

import (
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	DefaultMetricsPeriod        = 20 * time.Second
	DefaultWindowDuration       = 30 * time.Minute
	DefaultActiveBufferDuration = 5 * time.Minute
	DefaultAccessType           = telemetry.ServiceAccessPublic
)

type Telemetry struct {
	Enabled          bool
	BandwidthEnabled bool
	AccessType       telemetry.ServiceAccessType

	Whitelist            []peer.ID `json:",omitempty"`
	MetricsPeriod        string    `json:",omitempty"`
	WindowDuration       string    `json:",omitempty"`
	ActiveBufferDuration string    `json:",omitempty"`
	DebugListener        string    `json:",omitempty"`
}

func (t Telemetry) GetMetricsPeriod() time.Duration {
	return parseDurationOrDefault(t.MetricsPeriod, DefaultMetricsPeriod)
}

func (t Telemetry) GetWindowDuration() time.Duration {
	return parseDurationOrDefault(t.WindowDuration, DefaultWindowDuration)
}

func (t Telemetry) GetActiveBufferDuration() time.Duration {
	return parseDurationOrDefault(t.ActiveBufferDuration, DefaultActiveBufferDuration)
}

func parseDurationOrDefault(d string, def time.Duration) time.Duration {
	if dur, err := time.ParseDuration(d); err == nil {
		return dur
	} else {
		return def
	}
}
