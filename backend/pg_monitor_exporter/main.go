package pg_monitor_exporter

import (
	"time"

	"github.com/diogo464/ipfs-telemetry/backend"
	"github.com/diogo464/ipfs-telemetry/backend/crawler"
	"github.com/diogo464/ipfs-telemetry/backend/monitor"
	"github.com/nats-io/nats.go/jetstream"
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
	Name:        "pg_monitor_exporter",
	Description: "export monitor information to postgres",
	Flags: []cli.Flag{
		FlagRecreate,
	},
	Action: main,
}

func main(c *cli.Context) error {
	logger := backend.ServiceSetup(c, "pg-monitor-exporter")

	db := backend.PostgresClient(logger, c)
	nc := backend.NatsClient(logger, c)
	js := backend.NatsJetstream(logger, nc)
	defer db.Close(c.Context)

	recreate := c.Bool(FlagRecreate.Name)
	if recreate {
		db.Exec(c.Context, "DROP SCHEMA monitor CASCADE")
	}

	logger.Debug("running schema", zap.String("schema", schema))
	if _, err := db.Exec(c.Context, schema); recreate && err != nil {
		logger.Fatal("failed to execute schema", zap.Error(err))
	}

	startTime := time.Now().Add(-time.Second * 10)
	consumer, err := js.CreateConsumer(c.Context, monitor.Stream_Monitor, jetstream.ConsumerConfig{
		Description:   "crawler postgres exporter",
		FilterSubject: monitor.Subject_Active,
		DeliverPolicy: jetstream.DeliverByStartTimePolicy,
		OptStartTime:  &startTime,
		AckPolicy:     jetstream.AckNonePolicy,
	})
	backend.FatalOnError(logger, err, "failed to create monitor consumer")

	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		active := backend.NatsJetstreamDecodeJson[monitor.ActiveMessage](logger, msg)
		tx, err := db.Begin(c.Context)
		backend.FatalOnError(logger, err, "failed to start transaction")

		_, err = tx.Exec(c.Context, "DELETE FROM monitor.active")
		backend.FatalOnError(logger, err, "failed to delete current active peers")

		for _, peerId := range active.Peers {
			tx.Exec(c.Context, "INSERT INTO monitor.active(peer_id) VALUES ($1)", peerId.String())
		}

		err = tx.Commit(c.Context)
		backend.FatalOnError(logger, err, "failed to commit transaction")
	})
	backend.FatalOnError(logger, err, "failed to create nats consumer to crawler stream", zap.String("stream", crawler.StreamCrawler))
	<-cctx.Closed()

	return nil
}
