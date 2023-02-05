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
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	otel_host "go.opentelemetry.io/contrib/instrumentation/host"
	otel_runtime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	sdk_metric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var app = &cli.App{
	Name:   "example",
	Action: actionMain,
}

type ExampleEvent struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
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

	_, mp, err := telemetry.NewService(
		h,
		telemetry.WithServiceDebug(true),
		telemetry.WithServiceTcpListener("127.0.0.1:4000"),
		telemetry.WithServiceMetricsPeriod(time.Second*2),
		telemetry.WithServiceBandwidth(true),
		telemetry.WithServiceActiveBufferDuration(time.Second*5),
		telemetry.WithMeterProviderFactory(func(r sdk_metric.Reader) (metric.MeterProvider, error) {
			return sdk_metric.NewMeterProvider(
				sdk_metric.WithResource(resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceNameKey.String("example"),
				)),
				sdk_metric.WithReader(r),
			), nil
		}),
	)
	if err != nil {
		return err
	}

	global.SetMeterProvider(mp)
	otel_host.Start()
	otel_runtime.Start()

	tmp := mp
	m1 := tmp.TelemetryMeter("libp2p.io/host")

	m1.Property(
		"os",
		telemetry.NewPropertyValueString(runtime.GOOS),
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
			ev.Emit(&ExampleEvent{
				Field1: "field",
				Field2: c,
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
