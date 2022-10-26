package telemetry

import (
	"fmt"
	"time"

	"go.uber.org/atomic"
)

var ErrMetricAlreadyRegistered = fmt.Errorf("metric already registered")

var _ (MetricCollector) = (*Metric)(nil)
var _ (MetricCollector) = (*AutoMetric)(nil)

type CollectedMetric struct {
	Name  string
	Value float64
}

type MetricsObservations struct {
	Timestamp time.Time
	Metrics   map[string]float64
}

type MetricCollector interface {
	Collect(chan<- CollectedMetric)
}

type MetricConfig struct {
	Name         string
	DefaultValue float64
}

type Metric struct {
	name  string
	value *atomic.Float64
}

func NewMetric(config MetricConfig) *Metric {
	return &Metric{
		name:  config.Name,
		value: atomic.NewFloat64(config.DefaultValue),
	}
}

func (m *Metric) Inc() {
	m.value.Add(1.0)
}

func (m *Metric) Dec() {
	m.value.Sub(1.0)
}

func (m *Metric) Add(delta float64) {
	m.value.Add(delta)
}

func (m *Metric) Sub(delta float64) {
	m.value.Sub(delta)
}

func (m *Metric) Set(v float64) {
	m.value.Store(v)
}

// Collect implements MetricCollector
func (m *Metric) Collect(c chan<- CollectedMetric) {
	c <- CollectedMetric{Name: m.name, Value: m.value.Load()}
}

type AutoMetricConfig struct {
	Name      string
	Collector func() float64
}

type AutoMetric struct {
	name      string
	collector func() float64
}

func NewAutoMetric(config AutoMetricConfig) *AutoMetric {
	return &AutoMetric{
		name:      config.Name,
		collector: config.Collector,
	}
}

// Collect implements MetricCollector
func (m *AutoMetric) Collect(c chan<- CollectedMetric) {
	c <- CollectedMetric{Name: m.name, Value: m.collector()}
}
