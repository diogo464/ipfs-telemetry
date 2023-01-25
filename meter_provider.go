package telemetry

import (
	"sync"

	"go.opentelemetry.io/otel/metric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
)

var _ MeterProvider = (*serviceMeterProvider)(nil)

type MeterProvider interface {
	metric.MeterProvider

	TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter
}

type serviceMeterId struct {
	instrumentationName string
	config              metric.MeterConfig
}

type serviceMeterProvider struct {
	service        *Service
	meter_provider *sdk_metric.MeterProvider

	mu     sync.Mutex
	meters map[serviceMeterId]*serviceMeter
}

func newServiceMeterProvider(service *Service, meter_provider *sdk_metric.MeterProvider) *serviceMeterProvider {
	return &serviceMeterProvider{
		service:        service,
		meter_provider: meter_provider,

		meters: make(map[serviceMeterId]*serviceMeter),
	}
}

// Meter implements MeterProvider
func (mp *serviceMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return mp.TelemetryMeter(instrumentationName, opts...)
}

func (mp *serviceMeterProvider) TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter {
	cfg := metric.NewMeterConfig(opts...)
	meterId := serviceMeterId{instrumentationName: instrumentationName, config: cfg}

	mp.mu.Lock()
	defer mp.mu.Unlock()
	if meter, ok := mp.meters[meterId]; ok {
		return meter
	}

	meter := mp.meter_provider.Meter(instrumentationName, opts...)
	smeter := newServiceMeter(mp.service, instrumentationName, cfg, meter)
	mp.meters[meterId] = smeter

	return smeter
}

func DowncastMeterProvider(provider metric.MeterProvider) MeterProvider {
	if tprovider, ok := provider.(MeterProvider); ok {
		return tprovider
	} else {
		return NewNoopMeterProvider()
	}
}
