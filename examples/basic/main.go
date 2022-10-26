package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/diogo464/telemetry"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/urfave/cli/v2"
)

var (
	tPropertyHostOS = telemetry.NewStringProperty(telemetry.StringPropertyConfig{
		Name: "libp2p_host_os", Value: runtime.GOOS,
	})
	tPropertyHostArch = telemetry.NewStringProperty(telemetry.StringPropertyConfig{
		Name: "libp2p_host_arch", Value: runtime.GOARCH,
	})
	tPropertyHostNumCpu = telemetry.NewIntProperty(telemetry.IntPropertyConfig{
		Name: "libp2p_host_numcpu", Value: int64(runtime.NumCPU()),
	})

	tBitswapDiscoverySuccess = telemetry.NewMetric(telemetry.MetricConfig{Name: "libp2p_bitswap_discovery_success"})

	tKademliaHandler = telemetry.NewEvent(telemetry.EventConfig{
		Name: "libp2p_kad_handler",
	})

	tConnections = telemetry.NewSnapshot(telemetry.SnapshotConfig{
		Name:   "libp2p_network_connections",
		Period: time.Second * 5,
		Collector: func() (interface{}, error) {
			return struct {
				Connections int `json:"connections"`
			}{Connections: 5}, nil
		},
	})
)

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
		telemetry.WithServiceDefaultStreamOpts(
			telemetry.WithStreamSegmentLifetime(time.Second*60),
			telemetry.WithStreamActiveBufferLifetime(time.Second*5),
		),
	)
	if err != nil {
		return err
	}

	t.RegisterMetric(tBitswapDiscoverySuccess)
	t.RegisterEvent(tKademliaHandler)
	t.RegisterProperty(tPropertyHostOS)
	t.RegisterProperty(tPropertyHostArch)
	t.RegisterProperty(tPropertyHostNumCpu)
	t.RegisterSnapshot(tConnections)
	t.Start()

	go func() {
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				tBitswapDiscoverySuccess.Inc()
				tKademliaHandler.Emit(struct {
					Name    string `json:"name"`
					Handler uint64 `json:"handler"`
				}{Name: "FindNode", Handler: 51512})
			case <-c.Context.Done():
				break
			}
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
