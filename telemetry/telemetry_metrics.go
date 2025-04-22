package telemetry

import (
	"encoding/json"

	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type MetricDescriptor struct {
	Scope       string `json:"scope"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Metrics struct {
	OTLP []*mpb.ResourceMetrics `json:"otlp"`
}

func (m *Metrics) MarshalJSON() ([]byte, error) {
	// Use protojson to embed the field

	// Marshal the field
	marshaled := make([]json.RawMessage, len(m.OTLP))
	for i, v := range m.OTLP {
		b, err := protojson.MarshalOptions{
			UseProtoNames: true,
		}.Marshal(v)
		if err != nil {
			return nil, err
		}
		marshaled[i] = b
	}

	// Marshal the struct
	s := struct {
		OTLP []json.RawMessage `json:"otlp"`
	}{
		OTLP: marshaled,
	}

	return json.Marshal(s)
}

func (m *Metrics) UnmarshalJSON(b []byte) error {
	// Use protojson to embed the field

	// Unmarshal the struct
	s := struct {
		OTLP []json.RawMessage `json:"otlp"`
	}{}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// Unmarshal the field
	m.OTLP = make([]*mpb.ResourceMetrics, len(s.OTLP))
	for i, v := range s.OTLP {
		m.OTLP[i] = &mpb.ResourceMetrics{}
		if err := protojson.Unmarshal(v, m.OTLP[i]); err != nil {
			return err
		}
	}

	return nil
}
