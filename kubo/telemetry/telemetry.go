package telemetry

import (
	"context"
	"runtime"
	"time"

	"github.com/diogo464/telemetry"
	logging "github.com/ipfs/go-log"
	"github.com/ipfs/kubo/core"
	"github.com/ipfs/kubo/core/corerepo"
	"github.com/ipfs/kubo/telemetry/traceroute"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/multiformats/go-multiaddr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var log = logging.Logger("ipfs/telemetry")

type ProtocolStats struct {
	TotalIn  int64   `json:"total_in"`
	TotalOut int64   `json:"total_out"`
	RateIn   float64 `json:"rate_in"`
	RateOut  float64 `json:"rate_out"`
}

type Stream struct {
	Protocol string `json:"protocol"`
	// Timestamp of when the stream was opened
	Opened    int64  `json:"opened"`
	Direction string `json:"direction"`
}

type Traceroute struct {
	Target   peer.ID `json:"target"`
	Provider string  `json:"provider"`
	Output   []byte  `json:"output"`
}

type Connection struct {
	ID      peer.ID             `json:"id"`
	Addr    multiaddr.Multiaddr `json:"addr"`
	Latency int64               `json:"latency"`
	// Timestamp of when the connection was opened
	Opened  int64    `json:"opened"`
	Streams []Stream `json:"streams"`
}

func Start(node *core.IpfsNode) error {
	var t telemetry.MeterProvider
	gm := otel.GetMeterProvider()

	if tm, ok := gm.(telemetry.MeterProvider); ok {
		t = tm
	} else {
		t = telemetry.NewNoopMeterProvider()
	}

	if err := registerProperties(t); err != nil {
		return err
	}
	if err := registerNetworkCaptures(t, node); err != nil {
		return err
	}
	if err := registerStorageMetrics(t, node); err != nil {
		return err
	}
	if err := registerNetworkMetrics(t, node); err != nil {
		return err
	}
	if err := registerTraceroute(t, node); err != nil {
		return err
	}

	return nil
}

func registerProperties(t telemetry.MeterProvider) error {
	m := t.TelemetryMeter("libp2p.io/telemetry")

	m.Property(
		"process.runtime.os",
		telemetry.NewPropertyValueString(runtime.GOOS),
		metric.WithDescription("The operating system this node is running on. Obtained from runtime.GOOS"),
	)

	m.Property(
		"process.runtime.arch",
		telemetry.NewPropertyValueString(runtime.GOARCH),
		metric.WithDescription("The architecture this node is running on. Obtained from runtime.GOARCH"),
	)

	m.Property(
		"process.runtime.numcpu",
		telemetry.NewPropertyValueInteger(int64(runtime.NumCPU())),
		metric.WithDescription("The number of logical CPUs usable by the current process. Obtained from runtime.NumCPU"),
	)

	m.Property(
		"process.boottime",
		telemetry.NewPropertyValueInteger(time.Now().Unix()),
		metric.WithDescription("Boottime of this node in UNIX seconds"),
	)

	return nil
}

func registerNetworkCaptures(t telemetry.MeterProvider, node *core.IpfsNode) error {
	m := t.TelemetryMeter("libp2p.io/network")

	m.PeriodicEvent(
		context.TODO(),
		"libp2p.network.connections",
		time.Minute,
		func(_ context.Context, e telemetry.EventEmitter) error {
			networkConns := node.PeerHost.Network().Conns()
			connections := make([]Connection, 0, len(networkConns))

			for _, conn := range networkConns {
				streams := make([]Stream, 0, len(conn.GetStreams()))
				for _, stream := range conn.GetStreams() {
					streams = append(streams, Stream{
						Protocol:  string(stream.Protocol()),
						Opened:    stream.Stat().Opened.Unix(),
						Direction: stream.Stat().Direction.String(),
					})
				}
				connections = append(connections, Connection{
					ID:      conn.RemotePeer(),
					Addr:    conn.RemoteMultiaddr(),
					Latency: node.PeerHost.Network().Peerstore().LatencyEWMA(conn.RemotePeer()).Microseconds(),
					Opened:  conn.Stat().Opened.Unix(),
					Streams: streams,
				})
			}

			e.Emit(connections)
			return nil
		},
		metric.WithDescription("All current connections and streams of this node."),
	)

	m.PeriodicEvent(
		context.TODO(),
		"libp2p.network.addresses",
		2*time.Minute,
		func(_ context.Context, e telemetry.EventEmitter) error {
			e.Emit(node.PeerHost.Addrs())
			return nil
		},
		metric.WithDescription("The addresses the node is listening on"),
	)

	return nil
}

func registerStorageMetrics(t telemetry.MeterProvider, node *core.IpfsNode) error {
	var (
		err error

		storageUsed    metric.Int64ObservableUpDownCounter
		storageObjects metric.Int64ObservableUpDownCounter
		storageTotal   metric.Int64ObservableUpDownCounter
	)

	meter := t.Meter("libp2p.io/ipfs/storage")

	if storageUsed, err = meter.Int64ObservableUpDownCounter(
		"ipfs.storage.used",
		metric.WithUnit("By"),
		metric.WithDescription("Total number of bytes used by storage"),
	); err != nil {
		return err
	}

	if storageObjects, err = meter.Int64ObservableUpDownCounter(
		"ipfs.storage.objects",
		metric.WithUnit("1"),
		metric.WithDescription("Total number of objects in storage"),
	); err != nil {
		return err
	}

	if storageTotal, err = meter.Int64ObservableUpDownCounter(
		"ipfs.storage.total",
		metric.WithUnit("By"),
		metric.WithDescription("Total number of bytes avaible for storage"),
	); err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		stat, err := corerepo.RepoStat(ctx, node)
		if err != nil {
			log.Errorf("corerepo.RepoStat failed", "error", err)
			return err
		}

		obs.ObserveInt64(storageUsed, int64(stat.RepoSize))
		obs.ObserveInt64(storageObjects, int64(stat.NumObjects))
		obs.ObserveInt64(storageTotal, int64(stat.StorageMax))
		return nil
	}, storageUsed, storageObjects, storageTotal)
	if err != nil {
		return err
	}

	return nil
}

func registerNetworkMetrics(t telemetry.MeterProvider, node *core.IpfsNode) error {
	var (
		err error

		lowWater    metric.Int64ObservableUpDownCounter
		highWater   metric.Int64ObservableUpDownCounter
		connections metric.Int64ObservableUpDownCounter
		rateIn      metric.Int64ObservableUpDownCounter
		rateOut     metric.Int64ObservableUpDownCounter
		totalIn     metric.Int64ObservableCounter
		totalOut    metric.Int64ObservableCounter
	)

	m := t.Meter("libp2p.io/network")

	if lowWater, err = m.Int64ObservableUpDownCounter(
		"libp2p.network.low_water",
		metric.WithUnit("1"),
		metric.WithDescription("Network Low Water number of peers"),
	); err != nil {
		return err
	}

	if highWater, err = m.Int64ObservableUpDownCounter(
		"libp2p.network.high_water",
		metric.WithUnit("1"),
		metric.WithDescription("Network High Water number of peers"),
	); err != nil {
		return err
	}

	if connections, err = m.Int64ObservableUpDownCounter(
		"libp2p.network.connections",
		metric.WithUnit("1"),
		metric.WithDescription("Number of connections"),
	); err != nil {
		return err
	}

	if rateIn, err = m.Int64ObservableUpDownCounter(
		"libp2p.network.rate_in",
		metric.WithUnit("By"),
		metric.WithDescription("Network in rate in bytes per second"),
	); err != nil {
		return err
	}

	if rateOut, err = m.Int64ObservableUpDownCounter(
		"libp2p.network.rate_out",
		metric.WithUnit("By"),
		metric.WithDescription("Network out rate in bytes per second"),
	); err != nil {
		return err
	}

	if totalIn, err = m.Int64ObservableCounter(
		"libp2p.network.total_in",
		metric.WithUnit("By"),
		metric.WithDescription("Network total bytes in"),
	); err != nil {
		return err
	}

	if totalOut, err = m.Int64ObservableCounter(
		"libp2p.network.total_out",
		metric.WithUnit("By"),
		metric.WithDescription("Network total bytes out"),
	); err != nil {
		return err
	}

	_, err = m.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		reporter := node.Reporter
		cmgr := node.PeerHost.ConnManager().(*connmgr.BasicConnMgr)
		info := cmgr.GetInfo()

		obs.ObserveInt64(lowWater, int64(info.LowWater))
		obs.ObserveInt64(highWater, int64(info.HighWater))
		obs.ObserveInt64(connections, int64(info.ConnCount))

		bt := reporter.GetBandwidthTotals()
		obs.ObserveInt64(rateIn, int64(bt.RateIn))
		obs.ObserveInt64(rateOut, int64(bt.RateOut))
		obs.ObserveInt64(totalIn, bt.TotalIn)
		obs.ObserveInt64(totalOut, bt.TotalOut)

		for p, s := range node.Reporter.GetBandwidthByProtocol() {
			obs.ObserveInt64(rateIn, int64(s.RateIn), metric.WithAttributes(attribute.String("protocol", string(p))))
			obs.ObserveInt64(rateOut, int64(s.RateOut), metric.WithAttributes(attribute.String("protocol", string(p))))
			obs.ObserveInt64(totalIn, s.TotalIn, metric.WithAttributes(attribute.String("protocol", string(p))))
			obs.ObserveInt64(totalOut, s.TotalOut, metric.WithAttributes(attribute.String("protocol", string(p))))
		}

		return nil
	},
		lowWater,
		highWater,
		connections,
		rateIn,
		rateOut,
		totalIn,
		totalOut,
	)
	if err != nil {
		return err
	}

	return nil
}

func registerTraceroute(t telemetry.MeterProvider, node *core.IpfsNode) error {
	m := t.TelemetryMeter("libp2p.io/misc")

	picker := newPeerPicker(node.PeerHost)
	em := m.Event(
		"telemetry.misc.traceroute",
		metric.WithDescription("Traceroute"),
	)
	go func() {
		timeout := time.Second * 15

		for {
			time.Sleep(time.Second * 10)
			if pid, ok := picker.pick(); ok {
				addrinfo := node.PeerHost.Network().Peerstore().PeerInfo(pid)
				addr, err := getFirstPublicAddressFromMultiaddrs(addrinfo.Addrs)
				if err == nil {
					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					result, err := traceroute.Trace(ctx, addr.String())
					cancel()
					if err == nil {
						em.Emit(&Traceroute{
							Target:   pid,
							Provider: result.Provider,
							Output:   result.Output,
						})
					} else if err != traceroute.ErrNoProviderAvailable {
						log.Warn("Traceroute to ", addr, "failed with", err)
					}
				}
			}
		}
	}()

	return nil
}
