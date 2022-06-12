package datapoint

import (
	"time"
)

const WindowName = "window"

type Window struct {
	Timestamp       time.Time         `json:"timestamp"`
	WindowDuration  time.Duration     `json:"window_duration"`
	DatapointCount  map[string]uint32 `json:"datapoint.count"`
	DatapointMemory map[string]uint32 `json:"datapoint.memory"`
}
