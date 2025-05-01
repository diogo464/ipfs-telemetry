package backend

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
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

func NatsClient(logger *zap.Logger, c *cli.Context) *nats.Conn {
	natsUrl := c.String(Flag_NatsUrl.Name)
	logger.Info("connecting to nats", zap.String("url", natsUrl))
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		logger.Fatal("failed to connect to nats at "+natsUrl, zap.Error(err))
	}
	return nc
}

func NatsJetstream(logger *zap.Logger, nc *nats.Conn) jetstream.JetStream {
	js, err := jetstream.New(nc)
	if err != nil {
		logger.Fatal("failed to create jetstream context", zap.Error(err))
	}
	return js
}

func NatsPublishJson(logger *zap.Logger, nc *nats.Conn, subject string, value any) {
	serialized, err := json.Marshal(value)
	FatalOnError(logger, err, "failed to serialize json message to publish on nats", zap.String("subject", subject), zap.Any("value", value))
	err = nc.Publish(subject, serialized)
	FatalOnError(logger, err, "failed to publish message to nats", zap.String("subject", subject))
}

func NatsJetstreamDecodeJson[T any](logger *zap.Logger, msg jetstream.Msg) *T {
	value := new(T)
	err := json.Unmarshal(msg.Data(), value)
	FatalOnError(logger, err, "failed to decode nats jetstream json message",
		zap.String("subject", msg.Subject()), zap.Int("length", len(msg.Data())))
	return value
}

func NatsConsumer(ctx context.Context, logger *zap.Logger, js jetstream.JetStream, stream, consumer string) jetstream.Consumer {
	c, err := js.Consumer(ctx, stream, consumer)
	if err != nil {
		logger.Fatal("failed to create stream consumer", zap.Error(err))
	}
	return c
}

func PostgresClient(logger *zap.Logger, c *cli.Context) *pgx.Conn {
	databaseUrl := c.String(Flag_PostgresUrl.Name)
	conn, err := pgx.Connect(c.Context, databaseUrl)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.String("url", databaseUrl))
	}
	return conn
}

func FatalOnError(logger *zap.Logger, err error, msg string, fields ...zap.Field) {
	if err == nil {
		return
	}
	fields = append(fields, zap.Error(err))
	logger.Fatal(msg, fields...)
}
