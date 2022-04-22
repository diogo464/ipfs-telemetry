package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"git.d464.sh/adc/telemetry/pkg/crawler"
	pb "git.d464.sh/adc/telemetry/pkg/proto/crawler"
	"git.d464.sh/adc/telemetry/pkg/telemetry"
	"git.d464.sh/adc/telemetry/pkg/utils"
	"git.d464.sh/adc/telemetry/pkg/walker"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/mmcloughlin/geohash"
	"github.com/oschwald/geoip2-golang"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

var geodb *geoip2.Reader
var writer api.WriteAPI
var seen map[peer.ID]struct{} = make(map[peer.ID]struct{})

func main() {
	app := &cli.App{
		Name:        "crawler",
		Description: "discovery peers supporting the telemetry protocol",
		Action:      mainAction,
		Commands:    []*cli.Command{SubscribeCommand},
		Flags: []cli.Flag{
			FLAG_PROMETHEUS_ADDRESS,
			FLAG_INFLUXDB_ADDRESS,
			FLAG_INFLUXDB_TOKEN,
			FLAG_INFLUXDB_ORG,
			FLAG_INFLUXDB_BUCKET,
			FLAG_ADDRESS,
			FLAG_CONCURRENCY,
		},
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func testObserver(p *walker.Peer) {
	seen[p.ID] = struct{}{}
	for _, bucket := range p.Buckets {
		for _, info := range bucket {
			seen[info.ID] = struct{}{}
		}
	}

	public, err := utils.GetFirstPublicAddressFromMultiaddrs(p.Addresses)
	if err != nil {
		fmt.Println(err)
		return
	}

	city, err := geodb.City(public)
	if err != nil {
		fmt.Println(err)
	}

	var hasTelemetry string = "no"
	if utils.SliceAny(p.Protocols, func(t string) bool { return t == telemetry.ID_TELEMETRY }) {
		hasTelemetry = "yes"
	}
	gh := geohash.Encode(city.Location.Latitude, city.Location.Longitude)
	point := influxdb2.NewPointWithMeasurement("location").
		AddTag("peer", p.ID.Pretty()).
		AddField("geohash", gh).
		AddField("telemetry", hasTelemetry)
	writer.WritePoint(point)
}

func mainAction(c *cli.Context) error {
	gdb, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		return err
	}
	geodb = gdb

	client := influxdb2.NewClient(c.String(FLAG_INFLUXDB_ADDRESS.Name), c.String(FLAG_INFLUXDB_TOKEN.Name))
	defer client.Close()
	writer = client.WriteAPI(c.String(FLAG_INFLUXDB_ORG.Name), c.String(FLAG_INFLUXDB_BUCKET.Name))
	fmt.Println("ORG =", c.String(FLAG_INFLUXDB_ORG.Name), " BUCKET =", c.String(FLAG_INFLUXDB_BUCKET.Name))

	listener, err := net.Listen("tcp", c.String(FLAG_ADDRESS.Name))
	if err != nil {
		return err
	}
	grpc_server := grpc.NewServer()

	h, err := libp2p.New(libp2p.NoListenAddrs)
	if err != nil {
		return err
	}

	concurrency := c.Int(FLAG_CONCURRENCY.Name)
	fmt.Println("Crawler concurrency: ", concurrency)
	crlwr, err := crawler.NewCrawler(
		h,
		crawler.WithObserver(walker.NewPeerObserverFn(testObserver)),
		crawler.WithConcurrency(concurrency),
	)
	if err != nil {
		return err
	}

	pb.RegisterCrawlerServer(grpc_server, crlwr)
	go func() {
		err = grpc_server.Serve(listener)
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(c.String(FLAG_PROMETHEUS_ADDRESS.Name), nil))
	}()

	return crlwr.Run(c.Context)
}
