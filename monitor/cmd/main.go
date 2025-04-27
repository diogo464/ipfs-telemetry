package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tmonitor "github.com/diogo464/ipfs-telemetry/monitor"
	"github.com/diogo464/telemetry/monitor"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/multiformats/go-multiaddr"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
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

	_ "net/http/pprof"
)

func main() {
	app := &cli.App{
		Name:        "monitor",
		Description: "collect telemetry from ipfs nodes",
		Action:      mainAction,
		Flags: []cli.Flag{
			FLAG_PROMETHEUS_ADDRESS,
			FLAG_NATS_ENDPOINT,
			FLAG_MAX_FAILED_ATTEMPTS,
			FLAG_RETRY_INTERVAL,
			FLAG_COLLECT_ENABLED,
			FLAG_COLLECT_INTERVAL,
			FLAG_COLLECT_TIMEOUT,
			FLAG_BANDWIDTH_ENABLED,
			FLAG_BANDWIDTH_INTERVAL,
			FLAG_BANDWIDTH_TIMEOUT,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func mainAction(c *cli.Context) error {
	logger, _ := zap.NewProduction()

	prom_exporter, err := prometheus.New(prometheus.WithResourceAsConstantLabels(func(kv attribute.KeyValue) bool { return true }))
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(prom_exporter),
		metric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("monitor"),
			semconv.ServiceVersionKey.String("0.0.0"),
		)),
	)
	host.Start(host.WithMeterProvider(provider))
	runtime.Start(runtime.WithMeterProvider(provider))
	otel.SetMeterProvider(provider)

	monitorOptions := make([]monitor.Option, 0)

	if c.IsSet(FLAG_MAX_FAILED_ATTEMPTS.Name) {
		monitorOptions = append(monitorOptions, monitor.WithMaxFailedAttempts(c.Int(FLAG_MAX_FAILED_ATTEMPTS.Name)))
	}

	if c.IsSet(FLAG_RETRY_INTERVAL.Name) {
		monitorOptions = append(monitorOptions, monitor.WithRetryInterval(c.Duration(FLAG_RETRY_INTERVAL.Name)))
	}

	monitorOptions = append(monitorOptions, monitor.WithCollectEnabled(c.Bool(FLAG_COLLECT_ENABLED.Name)))

	if c.IsSet(FLAG_COLLECT_INTERVAL.Name) {
		monitorOptions = append(monitorOptions, monitor.WithCollectPeriod(c.Duration(FLAG_COLLECT_INTERVAL.Name)))
	}

	if c.IsSet(FLAG_COLLECT_TIMEOUT.Name) {
		monitorOptions = append(monitorOptions, monitor.WithCollectTimeout(c.Duration(FLAG_COLLECT_TIMEOUT.Name)))
	}

	monitorOptions = append(monitorOptions, monitor.WithBandwidthEnabled(c.Bool(FLAG_BANDWIDTH_ENABLED.Name)))

	if c.IsSet(FLAG_BANDWIDTH_INTERVAL.Name) {
		monitorOptions = append(monitorOptions, monitor.WithBandwidthPeriod(c.Duration(FLAG_BANDWIDTH_INTERVAL.Name)))
	}

	if c.IsSet(FLAG_BANDWIDTH_TIMEOUT.Name) {
		monitorOptions = append(monitorOptions, monitor.WithBandwidthTimeout(c.Duration(FLAG_BANDWIDTH_TIMEOUT.Name)))
	}

	natsAddress := c.String(FLAG_NATS_ENDPOINT.Name)
	logger.Info("connecting to nats at " + natsAddress)
	nc, err := nats.Connect(natsAddress)
	if err != nil {
		logger.Error("failed to connect to nats at "+natsAddress, zap.Error(err))
		return err
	}

	exporter := newNatsExporter(nc, logger.Named("exporter"))
	monitorOptions = append(monitorOptions, monitor.WithExporter(exporter))
	monitorOptions = append(monitorOptions, monitor.WithLogger(logger.Named("monitor")))
	monitorOptions = append(monitorOptions, monitor.WithMeterProvider(provider))

	limits := rcmgr.InfiniteLimits
	limiter := rcmgr.NewFixedLimiter(limits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return err
	}

	h, err := libp2p.New(libp2p.NoListenAddrs, libp2p.ResourceManager(rm))
	monitorOptions = append(monitorOptions, monitor.WithHost(h))

	mon, err := monitor.Start(c.Context, monitorOptions...)
	if err != nil {
		logger.Error("failed to start monitor", zap.Error(err))
		return err
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(c.String(FLAG_PROMETHEUS_ADDRESS.Name), nil))
	}()

	ch := make(chan *nats.Msg)
	sub, err := nc.ChanSubscribe(tmonitor.Subject_Discover, ch)
	if err != nil {
		return errors.Wrapf(err, "failed to subscribe to channel")
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-c.Context.Done():
			return nil
		case msg := <-ch:
			var discovery tmonitor.DiscoveryNotification
			err := json.Unmarshal(msg.Data, &discovery)
			if err != nil {
				logger.Error("failed to unmarshal discovery", zap.Error(err))
				continue
			}

			info, err := discoveryToAddrInfo(logger, &discovery)
			if err != nil {
				logger.Error("failed to create addr info from discovery", zap.Error(err))
				continue
			}

			mon.DiscoverWithAddr(c.Context, *info)
		}
	}
}

func discoveryToAddrInfo(logger *zap.Logger, discovery *tmonitor.DiscoveryNotification) (*peer.AddrInfo, error) {
	addrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range discovery.Addresses {
		prefix, comp := multiaddr.SplitLast(addr)
		if comp == nil {
			logger.Warn("failed to split multiaddr", zap.Any("addr", addr))
			continue
		}

		if comp.Protocol().Name == "p2p" {
			addrs = append(addrs, prefix)
		} else {
			addrs = append(addrs, addr)
		}
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("peer had not valid multiaddrs")
	}

	return &peer.AddrInfo{
		ID:    discovery.ID,
		Addrs: addrs,
	}, nil
}
