package telemetry

import "go.opentelemetry.io/otel/metric"

var _ MeterProvider = (*serviceMeterProvider)(nil)

type MeterProvider interface {
	metric.MeterProvider

	TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter
}

type serviceMeterProvider struct {
	service *Service
}

// Meter implements MeterProvider
func (mp *serviceMeterProvider) Meter(instrumentationName string, opts ...metric.MeterOption) metric.Meter {
	return mp.TelemetryMeter(instrumentationName, opts...)
}

func (mp *serviceMeterProvider) TelemetryMeter(instrumentationName string, opts ...metric.MeterOption) Meter {
	meter := mp.service.meter_provider.Meter(instrumentationName, opts...)
	config := metric.NewMeterConfig(opts...)
	return &serviceMeter{
		scope:   instrumentationName,
		config:  config,
		service: mp.service,
		meter:   meter,
	}
}

func DowncastMeterProvider(provider metric.MeterProvider) MeterProvider {
	if tprovider, ok := provider.(MeterProvider); ok {
		return tprovider
	} else {
		return NewNoopMeterProvider()
	}
}
