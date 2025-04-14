package telemetry

import (
	"sync"

	"go.opentelemetry.io/otel/metric"
)

var _ MeterProvider = (*serviceMeterProvider)(nil)

type MeterProvider interface {
	metric.MeterProvider

	TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter
}

type serviceMeterId struct {
	instrumentationName string
}

type serviceMeterProvider struct {
	metric.MeterProvider
	service *Service

	mu     sync.Mutex
	meters map[serviceMeterId]*serviceMeter
}

func newServiceMeterProvider(service *Service, meter_provider metric.MeterProvider) *serviceMeterProvider {
	return &serviceMeterProvider{
		MeterProvider: meter_provider,
		service:       service,

		meters: make(map[serviceMeterId]*serviceMeter),
	}
}

func (mp *serviceMeterProvider) TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter {
	// TODO: why am I saving the meters in a map? can't I just create a new one every time?
	cfg := metric.NewMeterConfig(opts...)
	meterId := serviceMeterId{instrumentationName: instrumentationName}

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if meter, ok := mp.meters[meterId]; ok {
		return meter
	}

	meter := mp.Meter(instrumentationName, opts...)
	smeter := newServiceMeter(mp.service, instrumentationName, cfg, meter)
	mp.meters[meterId] = smeter

	return smeter
}
