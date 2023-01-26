package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"

	otel_host "go.opentelemetry.io/contrib/instrumentation/host"
	otel_runtime "go.opentelemetry.io/contrib/instrumentation/runtime"
)

type HandlerTiming struct {
	Handler string
	Time1   uint64
	Time2   uint64
	Time3   uint64
}

var app = &cli.App{
	Name:   "node",
	Action: actionMain,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func actionMain(c *cli.Context) error {
	h, err := createHost()
	if err != nil {
		return err
	}

	t, err := telemetry.NewService(
		h,
		telemetry.WithServiceDebug(true),
		telemetry.WithServiceTcpListener("127.0.0.1:4000"),
		telemetry.WithServiceMetricsPeriod(time.Second*2),
		telemetry.WithServiceBandwidth(true),
		telemetry.WithServiceActiveBufferDuration(time.Second*5),
	)
	if err != nil {
		return err
	}

	mp := t.MeterProvider()
	global.SetMeterProvider(mp)
	otel_host.Start()
	otel_runtime.Start()

	tmp := mp
	m1 := tmp.TelemetryMeter("libp2p.io/host")

	m1.Property(
		"os",
		telemetry.PropertyValueString(runtime.GOOS),
		instrument.WithDescription("golang runtime.GOOS"))

	m2 := tmp.TelemetryMeter("libp2p.io/network")

	m2.PeriodicEvent(
		c.Context,
		"addresses",
		time.Second*5,
		func(ctx context.Context, e telemetry.EventEmitter) error {
			e.Emit(h.Addrs())
			return nil
		},
		instrument.WithDescription("some description"),
	)

	m3 := tmp.TelemetryMeter("libp2p.io/kad")
	ev := m3.Event("handler_timing", instrument.WithDescription("Handler timings"))

	go func() {
		c := 1
		for {
			time.Sleep(time.Millisecond * 2100)
			ev.Emit(&HandlerTiming{
				Handler: "handler",
				Time1:   uint64(c),
				Time2:   0,
				Time3:   0,
			})
			c += 1
		}
	}()

	fmt.Println(h.Network().ListenAddresses())
	fmt.Println(h.ID())

	<-c.Context.Done()

	return nil
}

func createHost() (host.Host, error) {
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/3001"))
	if err != nil {
		return nil, err
	}
	return h, nil
}
