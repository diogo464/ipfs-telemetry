package backend

import (
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

func ServiceSetup(c *cli.Context, name string) *zap.Logger {
	logger, _ := zap.NewProduction()

	prom_exporter, err := prometheus.New(prometheus.WithResourceAsConstantLabels(func(kv attribute.KeyValue) bool { return true }))
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(prom_exporter),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String("0.0.0"),
		)),
	)
	host.Start(host.WithMeterProvider(provider))
	runtime.Start(runtime.WithMeterProvider(provider))
	otel.SetMeterProvider(provider)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Fatal("failed to create prometheus http server", zap.Error(http.ListenAndServe(c.String(Flag_PrometheusAddress.Name), nil)))
	}()

	return logger
}

func CreateNatsClient(c *cli.Context, logger *zap.Logger) (*nats.Conn, error) {
	natsUrl := c.String(Flag_NatsUrl.Name)
	logger.Info("connecting to nats", zap.String("url", natsUrl))
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		logger.Error("failed to connect to nats at "+natsUrl, zap.Error(err))
		return nc, err
	}
	return nc, nil
}

func CreateNatsJetstream(nc *nats.Conn, logger *zap.Logger) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		logger.Error("failed to create jetstream context", zap.Error(err))
		return nil, err
	}
	return js, nil
}
