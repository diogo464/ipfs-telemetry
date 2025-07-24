package telemetry

import (
	"encoding/json"
	"time"

	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type MetricDescriptor struct {
	Scope       string `json:"scope"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Metrics struct {
	Timestamp time.Time            `json:"timestamp"`
	OTLP      *mpb.ResourceMetrics `json:"otlp"`
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	// Use protojson to embed the field

	// Marshal the field
	otlp, err := protojson.MarshalOptions{
		UseProtoNames: true,
	}.Marshal(m.OTLP)
	if err != nil {
		return nil, err
	}

	// Marshal the struct
	s := struct {
		Timestamp time.Time       `json:"timestamp"`
		OTLP      json.RawMessage `json:"otlp"`
	}{
		Timestamp: m.Timestamp,
		OTLP:      otlp,
	}

	return json.Marshal(s)
}

func (m *Metrics) UnmarshalJSON(b []byte) error {
	// Use protojson to embed the field

	// Unmarshal the struct
	s := struct {
		Timestamp time.Time       `json:"timestamp"`
		OTLP      json.RawMessage `json:"otlp"`
	}{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Unmarshal the field
	m.Timestamp = s.Timestamp
	m.OTLP = &mpb.ResourceMetrics{}
	if err := protojson.Unmarshal(s.OTLP, m.OTLP); err != nil {
		return err
	}

	return nil
}
