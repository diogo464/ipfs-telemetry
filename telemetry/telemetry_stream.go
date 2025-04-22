package telemetry

import (
	"github.com/diogo464/telemetry/internal/pb"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

var (
	_ (StreamType) = (*StreamTypeMetric)(nil)
	_ (StreamType) = (*StreamTypeEvent)(nil)
)

type StreamId uint32

type StreamType interface {
	sealed()
}

type StreamTypeMetric struct{}

// sealed implements StreamType
func (*StreamTypeMetric) sealed() {
}

type StreamTypeEvent struct {
	Scope       instrumentation.Scope `json:"scope"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
}

// sealed implements StreamType
func (*StreamTypeEvent) sealed() {
}

type StreamDescriptor struct {
	ID   StreamId   `json:"id"`
	Type StreamType `json:"type"`
}

func streamDescriptorFromPb(d *pb.StreamDescriptor) StreamDescriptor {
	var t StreamType
	switch d.StreamType.Type.(type) {
	case *pb.StreamType_Event:
		t = &StreamTypeEvent{
			Scope: instrumentation.Scope{
				Name:    d.StreamType.GetEvent().Scope.Name,
				Version: d.StreamType.GetEvent().Scope.Version,
			},
			Name:        d.StreamType.GetEvent().Name,
			Description: d.StreamType.GetEvent().Description,
		}
	case *pb.StreamType_Metric:
		t = &StreamTypeMetric{}
	}
	return StreamDescriptor{
		ID:   StreamId(d.StreamId),
		Type: t,
	}
}
