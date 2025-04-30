package crawler

import (
	"github.com/diogo464/ipfs-telemetry/backend"
	"github.com/diogo464/telemetry/crawler"
	"github.com/diogo464/telemetry/walker"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/otel"
)

var Command *cli.Command = &cli.Command{
	Name:        "crawler",
	Description: "crawler service",
	Flags: []cli.Flag{
		FLAG_CONCURRENCY,
		FLAG_CONNECT_TIMEOUT,
		FLAG_REQUEST_TIMEOUT,
		FLAG_INTERVAL,
	},
	Action: main,
}

func main(c *cli.Context) error {
	logger := backend.ServiceSetup(c, "crawler")

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

	url := c.String(backend.Flag_NatsUrl.Name)
	natsObserver, err := newNatsObserver(logger.Named("nats-observer"), url)
	if err != nil {
		return err
	}

	logger.Info("creating crawler")
	crlwr, err := crawler.NewCrawler(
		crawler.WithWalkerObserver(newLoggerObserver(logger)),
		crawler.WithObserver(natsObserver),
		crawler.WithWalkerOption(walkerOpts...),
		crawler.WithLogger(logger.Named("crawler")),
		crawler.WithMeterProvider(otel.GetMeterProvider()),
	)
	if err != nil {
		return err
	}

	logger.Info("starting crawler")
	return crlwr.Run(c.Context)
}
