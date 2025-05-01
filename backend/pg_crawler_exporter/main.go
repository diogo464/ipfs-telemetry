package pg_crawler_exporter

import (
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/diogo464/ipfs-telemetry/backend"
	"github.com/diogo464/ipfs-telemetry/backend/crawler"
	"github.com/diogo464/telemetry/walker"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/oschwald/geoip2-golang"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	_ "embed"
)

//go:embed schema.sql
var schema string

var FlagRecreate *cli.BoolFlag = &cli.BoolFlag{
	Name:  "recreate",
	Usage: "recreate the postgres database schema",
	Value: false,
}

var Command *cli.Command = &cli.Command{
	Name:        "pg_crawler_exporter",
	Description: "export crawler information to postgres",
	Flags: []cli.Flag{
		FlagRecreate,
	},
	Action: main,
}

const (
	ProtocolTcp = "tcp"
	ProtocolUdp = "udp"
)

type PublicAddress struct {
	Ip       net.IP
	Port     uint16
	Protocol string
}

func main(c *cli.Context) error {
	logger := backend.ServiceSetup(c, "pg-crawler-exporter")

	conn := backend.PostgresClient(logger, c)
	nc := backend.NatsClient(logger, c)
	js := backend.NatsJetstream(logger, nc)

	defer conn.Close(c.Context)

	cityDbPath := "./GeoLite2-City.mmdb"
	asnDbPath := "./GeoLite2-ASN.mmdb"
	cityDb, err := geoip2.Open(cityDbPath)
	if err != nil {
		logger.Fatal("failed to open geolite City database", zap.String("path", cityDbPath), zap.Error(err))
	}
	asnDb, err := geoip2.Open(asnDbPath)
	if err != nil {
		logger.Fatal("failed to open geolite ASN database", zap.String("path", asnDbPath), zap.Error(err))
	}

	recreate := c.Bool(FlagRecreate.Name)
	if recreate {
		conn.Exec(c.Context, "DROP SCHEMA crawler CASCADE")
	}

	logger.Debug("running schema", zap.String("schema", schema))
	if _, err := conn.Exec(c.Context, schema); recreate && err != nil {
		logger.Fatal("failed to execute schema", zap.Error(err))
	}

	consumer, err := js.CreateConsumer(c.Context, crawler.StreamCrawler, jetstream.ConsumerConfig{
		Description: "crawler postgres exporter",
		AckPolicy:   jetstream.AckNonePolicy,
	})
	if err != nil {
		logger.Fatal("failed to create crawler consumer", zap.Error(err))
	}

	peerIds := make(map[peer.ID]uint64)
	crawlBegin := time.Time{}
	crawlInProgress := false
	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		metadata, _ := msg.Metadata()
		logger.Info("received crawler message", zap.Int("size", len(msg.Data())), zap.Uint64("seqn", metadata.Sequence.Stream))
		cmsg := crawler.NatsMessage{}
		if err := json.Unmarshal(msg.Data(), &cmsg); err != nil {
			logger.Error("failed to decode crawler nats message", zap.Error(err))
			return
		}

		switch cmsg.Kind {
		case crawler.KindCrawlBegin:
			peerIds = make(map[peer.ID]uint64)
			logger.Info("crawl started")

			if _, err := conn.Exec(c.Context, "DELETE FROM crawler.peer WHERE crawl = NULL"); err != nil {
				logger.Fatal("failed to remove existing peers without associated crawl", zap.Error(err))
			}

			crawlBegin = cmsg.Timestamp
			crawlInProgress = true
		case crawler.KindCrawlEnd:
			peerIds = make(map[peer.ID]uint64)
			logger.Info("crawl ended")

			if crawlInProgress {
				tx, err := conn.Begin(c.Context)
				if err != nil {
					logger.Fatal("failed to create transaction", zap.Error(err))
				}

				row := tx.QueryRow(c.Context, "INSERT INTO crawler.crawl(timestamp_begin, timestamp_end) VALUES($1, $2) RETURNING id", crawlBegin, cmsg.Timestamp)

				var id int
				if err := row.Scan(&id); err != nil {
					logger.Fatal("failed to obtain id of newly created crawl", zap.Error(err))
				}

				if _, err := tx.Exec(c.Context, "UPDATE crawler.peer SET crawl = $1 WHERE crawl IS NULL", id); err != nil {
					logger.Fatal("failed to update crawl id on existing peers", zap.Int("id", id), zap.Error(err))
				}

				if err := tx.Commit(c.Context); err != nil {
					logger.Fatal("failed to commit transaction", zap.Error(err))
				}
			}

			crawlInProgress = false
		case crawler.KindPeer:
			p := cmsg.Peer
			addrs := make([]string, len(p.Addresses))
			for i, maddr := range p.Addresses {
				addrs[i] = maddr.String()
			}
			protocols := make([]string, len(p.Protocols))
			for i, proto := range p.Protocols {
				protocols[i] = string(proto)
			}

			var ip *string
			var country *string
			var city *string
			var asn *uint
			var asnOrg *string
			var latitude *float64
			var longitude *float64
			if pip, ok := ExtractFirstPublicIp(p.Addresses); ok {
				if gasn, err := asnDb.ASN(pip); err == nil {
					asn = &gasn.AutonomousSystemNumber
					asnOrg = &gasn.AutonomousSystemOrganization
				} else {
					logger.Warn("failed to obtain ip asn information", zap.String("ip", pip.String()), zap.Error(err))
				}
				if gcity, err := cityDb.City(pip); err == nil {
					if countryName, ok := gcity.Country.Names["en"]; ok {
						country = &countryName
					}
					if cityName, ok := gcity.City.Names["en"]; ok {
						city = &cityName
					}
					latitude = &gcity.Location.Latitude
					longitude = &gcity.Location.Longitude
				} else {
					logger.Warn("failed to obtain ip city information", zap.String("ip", pip.String()), zap.Error(err))
				}

				ip = new(string)
				*ip = pip.String()
			}

			if s, ok := peerIds[p.ID]; ok {
				logger.Fatal("peer id already exists", zap.String("id", p.ID.String()), zap.Uint64("seqn", s))
			}
			peerIds[p.ID] = metadata.Sequence.Stream
			_, err := conn.Exec(c.Context,
				"INSERT INTO crawler.peer(timestamp, peer_id, agent, addresses, protocols, dht_entries, ip, asn, asn_org, country, city, latitude, longitude) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
				cmsg.Timestamp, p.ID.String(), p.Agent, addrs, protocols, len(p.Buckets), ip, asn, asnOrg, country, city, latitude, longitude)
			if err != nil {
				logger.Fatal("failed to insert peer entry into database", zap.Error(err))
			}
		default:
			logger.Fatal("unknown crawler message kind", zap.String("kind", cmsg.Kind))
		}
	})
	if err != nil {
		logger.Fatal("failed to create nats consumer to crawler stream", zap.String("stream", crawler.StreamCrawler), zap.Error(err))
	}
	<-cctx.Closed()

	return nil
}

func containsProtocol(addr multiaddr.Multiaddr, name string) bool {
	for _, c := range addr {
		if c.Protocol().Name == name {
			return true
		}
	}
	return false
}

func ExtractFirstPublicIp(maddrs []multiaddr.Multiaddr) (net.IP, bool) {
	for _, maddr := range maddrs {
		if containsProtocol(maddr, "p2p-circuit") {
			continue
		}

		ip := net.ParseIP(maddr[0].Value())
		if ip == nil {
			continue
		}

		if ip.IsPrivate() {
			continue
		}

		return ip, true
	}
	return nil, false
}

func ExtractPublicAddresses(maddrs []multiaddr.Multiaddr) []PublicAddress {
	addrs := make([]PublicAddress, 0)
	for _, maddr := range maddrs {
		if containsProtocol(maddr, "p2p-circuit") {
			continue
		}

		if len(maddr) < 4 {
			continue
		}

		ip := net.ParseIP(maddr[0].Value())
		if ip == nil {
			continue
		}
		protocol := maddr[1].Protocol().Name
		port, err := strconv.ParseUint(maddr[1].Value(), 10, 16)
		if err != nil {
			continue
		}

		addrs = append(addrs, PublicAddress{
			Ip:       ip,
			Port:     uint16(port),
			Protocol: protocol,
		})
	}
	return addrs
}

func foo(db *geoip2.Reader, p *walker.Peer) {
	// for _, addr := range p.Addresses {
	// }
}
