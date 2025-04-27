package main

import (
	"log"
	"net/http"
	"os"

	"github.com/diogo464/telemetry/crawler"
	"github.com/diogo464/telemetry/walker"
	logging "github.com/ipfs/go-log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap"
)

func main() {
	app := &cli.App{
		Name:        "crawler",
		Description: "discovery peers supporting the telemetry protocol",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_PROMETHEUS_ADDRESS,
			FLAG_NATS_URL,
			FLAG_CONCURRENCY,
			FLAG_CONNECT_TIMEOUT,
			FLAG_REQUEST_TIMEOUT,
			FLAG_INTERVAL,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	logger, _ := zap.NewProduction()

	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("crawler"),
			semconv.ServiceVersionKey.String("0.0.0"),
		)),
	)

	if v, ok := os.LookupEnv("GOLOG_LOG_LEVEL"); ok && v == "debug" {
		logging.SetAllLoggers(logging.LevelDebug)
	}

	walkerOpts := []walker.Option{}
	if c.IsSet(FLAG_CONCURRENCY.Name) {
		walkerOpts = append(walkerOpts, walker.WithConcurrency(uint(c.Int(FLAG_CONCURRENCY.Name))))
	}
	if c.IsSet(FLAG_CONNECT_TIMEOUT.Name) {
		walkerOpts = append(walkerOpts, walker.WithConnectTimeout(c.Duration(FLAG_CONNECT_TIMEOUT.Name)))
	}
	if c.IsSet(FLAG_REQUEST_TIMEOUT.Name) {
		walkerOpts = append(walkerOpts, walker.WithRequestTimeout(c.Duration(FLAG_REQUEST_TIMEOUT.Name)))
	}
	if c.IsSet(FLAG_INTERVAL.Name) {
		walkerOpts = append(walkerOpts, walker.WithInterval(c.Duration(FLAG_INTERVAL.Name)))
	}

	url := c.String(FLAG_NATS_URL.Name)
	natsObserver, err := newNatsObserver(logger.Named("nats"), url)
	if err != nil {
		return err
	}

	logger.Info("creating crawler")
	crlwr, err := crawler.NewCrawler(
		crawler.WithWalkerObserver(newLoggerObserver(logger)),
		crawler.WithObserver(natsObserver),
		crawler.WithWalkerOption(walkerOpts...),
		crawler.WithLogger(logger.Named("crawler")),
		crawler.WithMeterProvider(provider),
	)
	if err != nil {
		return err
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(c.String(FLAG_PROMETHEUS_ADDRESS.Name), nil))
	}()

	logger.Info("starting crawler")
	return crlwr.Run(c.Context)
}
