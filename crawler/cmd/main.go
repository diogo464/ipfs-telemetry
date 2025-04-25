package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	backend_crawler "github.com/diogo464/ipfs-telemetry/crawler"
	"github.com/diogo464/telemetry/crawler"
	"github.com/diogo464/telemetry/walker"
	logging "github.com/ipfs/go-log"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.uber.org/zap"
)

var _ walker.Observer = (*observer)(nil)
var _ walker.Observer = (*natsObserver)(nil)
var _ walker.Observer = (*httpObserver)(nil)

func main() {
	app := &cli.App{
		Name:        "crawler",
		Description: "discovery peers supporting the telemetry protocol",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_PROMETHEUS_ADDRESS,
			FLAG_NATS_URL,
			FLAG_NATS_SUBJECT,
			FLAG_OUTPUT,
			FLAG_HTTP_URL,
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

func walkerPeerToDiscovery(c *walker.Peer) backend_crawler.DiscoveryNotification {
	return backend_crawler.DiscoveryNotification{
		ID:        c.ID,
		Addresses: c.Addresses,
	}
}

func walkerPeerToDiscoveryMarshaled(c *walker.Peer) ([]byte, error) {
	d := walkerPeerToDiscovery(c)
	m, err := json.Marshal(d)
	return m, err
}

type observer struct {
	l *zap.Logger
	o walker.Observer
}

// ObserveError implements walker.Observer
func (o *observer) ObserveError(e *walker.Error) {
	o.l.Warn("error", zap.String("peer", e.ID.String()), zap.Error(e.Err))
	o.o.ObserveError(e)
}

// ObservePeer implements walker.Observer
func (o *observer) ObservePeer(c *walker.Peer) {
	o.l.Info("observing peer", zap.String("peer", c.ID.String()))
	o.o.ObservePeer(c)
}

type natsObserver struct {
	l       *zap.Logger
	nc      *nats.Conn
	subject string
}

func newNatsObserver(l *zap.Logger, natsUrl string, subject string) (*natsObserver, error) {
	l.Info("connecting to nats at " + natsUrl)
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		l.Error("failed to connect to nats at "+natsUrl, zap.Error(err))
		return nil, err
	}
	return &natsObserver{
		l:       l,
		nc:      nc,
		subject: subject,
	}, nil
}

// ObserveError implements walker.Observer
func (*natsObserver) ObserveError(*walker.Error) {
}

// ObservePeer implements walker.Observer
func (o *natsObserver) ObservePeer(c *walker.Peer) {
	if m, err := walkerPeerToDiscoveryMarshaled(c); err == nil {
		if err := o.nc.Publish(o.subject, m); err != nil {
			o.l.Error("failed to publish discovery message", zap.String("subject", o.subject), zap.Error(err))
		}
	} else {
		o.l.Error("failed to marshal discovery", zap.Error(err))
	}
}

type httpObserver struct {
	l      *zap.Logger
	apiUrl string
}

func newHttpObserver(l *zap.Logger, apiUrl string) *httpObserver {
	return &httpObserver{
		l:      l,
		apiUrl: apiUrl,
	}
}

// ObserveError implements walker.Observer
func (*httpObserver) ObserveError(*walker.Error) {
}

// ObservePeer implements walker.Observer
func (o *httpObserver) ObservePeer(c *walker.Peer) {
	if m, err := walkerPeerToDiscoveryMarshaled(c); err == nil {
		postUrl := o.apiUrl + "/discovery"
		if _, err := http.Post(postUrl, "application/json", bytes.NewReader(m)); err != nil {
			o.l.Error("failed to POST discovery", zap.String("url", postUrl), zap.Error(err))
		}
	} else {
		o.l.Error("failed to marshal discovery", zap.Error(err))
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

	outputMethod := c.String(FLAG_OUTPUT.Name)
	var outputObserver walker.Observer
	if outputMethod == "http" {
		baseUrl := c.String(FLAG_HTTP_URL.Name)
		outputObserver = newHttpObserver(logger.Named("http"), baseUrl)
	} else if outputMethod == "nats" {
		url := c.String(FLAG_NATS_URL.Name)
		subject := c.String(FLAG_NATS_SUBJECT.Name)
		if natsObserver, err := newNatsObserver(logger.Named("nats"), url, subject); err == nil {
			outputObserver = natsObserver
		} else {
			return err
		}
	} else {
		log.Fatal("unknown output method", zap.String("output", outputMethod))
	}

	observer := &observer{
		l: logger,
		o: outputObserver,
	}

	logger.Info("creating crawler")
	crlwr, err := crawler.NewCrawler(
		crawler.WithObserver(observer),
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
