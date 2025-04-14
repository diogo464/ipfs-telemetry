package libp2p

import (
	"fmt"

	"github.com/diogo464/telemetry"
	version "github.com/ipfs/kubo"
	ipfs_config "github.com/ipfs/kubo/config"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

type HostOption func(id peer.ID, ps peerstore.Peerstore, cfg ipfs_config.Telemetry, options ...libp2p.Option) (host.Host, error)

var DefaultHostOption HostOption = constructPeerHost

// isolates the complex initialization steps
func constructPeerHost(id peer.ID, ps peerstore.Peerstore, cfg ipfs_config.Telemetry, options ...libp2p.Option) (host.Host, error) {
	pkey := ps.PrivKey(id)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.String())
	}
	options = append([]libp2p.Option{libp2p.Identity(pkey), libp2p.Peerstore(ps)}, options...)

	telemetryConstructor := func(h host.Host) error {
		opts := []telemetry.ServiceOption{
			telemetry.WithServiceMetricsPeriod(cfg.GetMetricsPeriod()),
			telemetry.WithServiceBandwidth(cfg.BandwidthEnabled),
			telemetry.WithServiceActiveBufferDuration(cfg.GetActiveBufferDuration()),
			telemetry.WithServiceWindowDuration(cfg.GetWindowDuration()),
			telemetry.WithServiceAccessType(cfg.AccessType),
			telemetry.WithServiceAccessWhitelist(cfg.Whitelist...),

			// MeterProvider Factory
			telemetry.WithMeterProviderFactory(func(r sdk_metric.Reader) (metric.MeterProvider, error) {
				return sdk_metric.NewMeterProvider(
					sdk_metric.WithResource(resource.NewWithAttributes(
						semconv.SchemaURL,
						semconv.ServiceNameKey.String("ipfs"),
						semconv.ServiceVersionKey.String(version.CurrentVersionNumber),
						attribute.String("peerid", id.String()),
					)),
					sdk_metric.WithReader(r),
				), nil
			}),
		}

		if len(cfg.DebugListener) > 0 {
			opts = append(opts, telemetry.WithServiceTcpListener(cfg.DebugListener))
		}

		_, mp, err := telemetry.NewService(h, opts...)
		if err != nil {
			return err
		}

		otel.SetMeterProvider(mp)

		return nil
	}

	options = append(options, func(cfg *config.Config) error {
		rctor := cfg.Routing
		if rctor != nil {
			cfg.Routing = func(h host.Host) (routing.PeerRouting, error) {
				if err := telemetryConstructor(h); err != nil {
					return nil, err
				}
				return rctor(h)
			}
		}
		return nil
	})

	h, err := libp2p.New(options...)
	if err != nil {
		return nil, err
	}

	if _, ok := otel.GetMeterProvider().(telemetry.MeterProvider); !ok {
		if err := telemetryConstructor(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}
