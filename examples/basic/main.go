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
		telemetry.WithServiceDefaultStreamOpts(
			telemetry.WithStreamSegmentLifetime(time.Second*60),
			telemetry.WithStreamActiveBufferLifetime(time.Second*5),
		),
	)
	if err != nil {
		return err
	}

	telemetry.SetGlobalTelemetry(t)
	global.SetMeterProvider(t)
	otel_host.Start()
	otel_runtime.Start()

	t.Property(telemetry.PropertyConfig{
		Name:        "libp2p_host_os",
		Description: "golang runtime.GOOS",
		Value:       telemetry.NewPropertyValueString(runtime.GOOS),
	})
	t.Property(telemetry.PropertyConfig{
		Name:        "libp2p_host_arch",
		Description: "golang runtime.GOARCH",
		Value:       telemetry.NewPropertyValueString(runtime.GOARCH),
	})
	t.Property(telemetry.PropertyConfig{
		Name:        "libp2p_host_numcpu",
		Description: "golang runtime.NumCPU()",
		Value:       telemetry.NewPropertyValueInteger(int64(runtime.NumCPU())),
	})
	t.Capture(telemetry.CaptureConfig{
		Name:        "libp2p_network_addrs",
		Description: "stuff and things",
		Callback: func(context.Context) (interface{}, error) {
			return h.Addrs(), nil
		},
		Interval: time.Second * 5,
	})
	ev := t.Event(telemetry.EventConfig{
		Name:        "libp2p_kad_handler",
		Description: "Handler timings",
	})

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
