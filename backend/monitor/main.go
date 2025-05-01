package monitor

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/diogo464/ipfs-telemetry/backend"
	"github.com/diogo464/telemetry/monitor"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/multiformats/go-multiaddr"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

var Command *cli.Command = &cli.Command{
	Name:        "monitor",
	Description: "monitor service",
	Flags: []cli.Flag{
		FLAG_MAX_FAILED_ATTEMPTS,
		FLAG_RETRY_INTERVAL,
		FLAG_COLLECT_ENABLED,
		FLAG_COLLECT_INTERVAL,
		FLAG_COLLECT_TIMEOUT,
		FLAG_BANDWIDTH_ENABLED,
		FLAG_BANDWIDTH_INTERVAL,
		FLAG_BANDWIDTH_TIMEOUT,
	},
	Action: main,
}

func main(c *cli.Context) error {
	logger := backend.ServiceSetup(c, "monitor")

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

	nc := backend.NatsClient(logger, c)
	js := backend.NatsJetstream(logger, nc)

	exporter := newExporter(nc, logger.Named("exporter"))
	monitorOptions = append(monitorOptions, monitor.WithExporter(exporter))
	monitorOptions = append(monitorOptions, monitor.WithLogger(logger.Named("telemetry.monitor")))
	monitorOptions = append(monitorOptions, monitor.WithMeterProvider(otel.GetMeterProvider()))

	limits := rcmgr.InfiniteLimits
	limiter := rcmgr.NewFixedLimiter(limits)
	rm, err := rcmgr.NewResourceManager(limiter)
	backend.FatalOnError(logger, err, "failed to create resource manager")

	h, err := libp2p.New(libp2p.NoListenAddrs, libp2p.ResourceManager(rm))
	monitorOptions = append(monitorOptions, monitor.WithHost(h))

	mon, err := monitor.Start(c.Context, monitorOptions...)
	backend.FatalOnError(logger, err, "failed to start monitor")

	go func() {
		for {
			ticker := time.NewTicker(time.Second * 5)
			select {
			case <-ticker.C:
				backend.NatsPublishJson(logger, nc, Subject_Active, &ActiveMessage{
					Peers: mon.GetActivePeers(),
				})
			case <-c.Context.Done():
				return
			}
		}
	}()

	since := time.Now().Add(-time.Hour)
	cfg := jetstream.ConsumerConfig{
		Description:   "monitor service discover consumer",
		DeliverPolicy: jetstream.DeliverByStartTimePolicy,
		OptStartTime:  &since,
		FilterSubject: Subject_Discover,
		AckPolicy:     jetstream.AckNonePolicy,
	}
	consumer, err := js.CreateConsumer(c.Context, "monitor", cfg)
	if err != nil {
		logger.Error("failed to create jetstream consumer", zap.Error(err))
		return err
	}
	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		var discovery DiscoveryMessage
		err := json.Unmarshal(msg.Data(), &discovery)
		if err != nil {
			logger.Error("failed to unmarshal discovery", zap.Error(err))
			return
		}

		info, err := discoveryToAddrInfo(logger, &discovery)
		if err != nil {
			logger.Error("failed to create addr info from discovery", zap.Error(err))
			return
		}

		mon.DiscoverWithAddr(c.Context, *info)
	})

	if err != nil {
		logger.Error("failed to consume jetstream messages", zap.Error(err))
		return err
	}

	<-cctx.Closed()
	return nil
}

func discoveryToAddrInfo(logger *zap.Logger, discovery *DiscoveryMessage) (*peer.AddrInfo, error) {
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
