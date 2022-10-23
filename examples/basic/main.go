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
	tOpts = []telemetry.CollectorOption{telemetry.WithCollectorPeriod(time.Second)}

	tPropertyHostOS     = telemetry.NewConstStrProperty("libp2p_host_os", runtime.GOOS)
	tPropertyHostArch   = telemetry.NewConstStrProperty("libp2p_host_arch", runtime.GOARCH)
	tPropertyHostNumCpu = telemetry.NewConstInt64Property("libp2p_host_numcpu", int64(runtime.NumCPU()))

	tBitswapDiscoverySuccess = telemetry.NewInt64Metric("libp2p_bitswap_discovery_success", tOpts...)
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
		telemetry.WithServiceDefaultStreamOpts(telemetry.WithStreamSegmentLifetime(time.Second*5)),
		telemetry.WithServiceDebug(true),
		telemetry.WithTcpListener("127.0.0.1:4000"),
		telemetry.WithServiceDefaultStreamOpts(telemetry.WithStreamActiveBufferLifetime(time.Second*5)),
	)
	if err != nil {
		return err
	}

	t.Register(tBitswapDiscoverySuccess)
	t.RegisterProperty(tPropertyHostOS)
	t.RegisterProperty(tPropertyHostArch)
	t.RegisterProperty(tPropertyHostNumCpu)

	go func() {
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				tBitswapDiscoverySuccess.Inc()
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
