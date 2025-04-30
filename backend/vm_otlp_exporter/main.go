package vm_otlp_exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/diogo464/ipfs-telemetry/backend"
	"github.com/diogo464/ipfs-telemetry/backend/monitor"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/urfave/cli/v2"
	v1_service "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"

	"google.golang.org/protobuf/proto"
)

var Command *cli.Command = &cli.Command{
	Name:        "vm-otlp-exporter",
	Description: "export OTLP metrics to VictoriaMetrics",
	Flags:       []cli.Flag{},
	Action:      main,
}

func main(c *cli.Context) error {
	logger := backend.ServiceSetup(c, "vm-otlp-exporter")

	client, err := backend.CreateNatsClient(c, logger)
	if err != nil {
		return err
	}

	js, err := backend.CreateNatsJetstream(client, logger)
	if err != nil {
		return err
	}

	stream, err := js.Stream(c.Context, "monitor")
	if err != nil {
		logger.Error("failed to create monitor stream", zap.Error(err))
		return err
	}

	consumer, err := stream.Consumer(c.Context, "monitor-vm-otlp-exporter")
	if err != nil {
		logger.Error("failed to create stream consumer", zap.Error(err))
		return err
	}

	vmExportUrl := fmt.Sprintf("%s/opentelemetry/v1/metrics", c.String(backend.Flag_VmUrl.Name))
	logger.Info("starting victoria metrics otlp exporter", zap.String("export-url", vmExportUrl))

	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		meta, _ := msg.Metadata()
		logger.Info("processing2 message", zap.Uint64("seqn", meta.Sequence.Stream))

		export := new(monitor.Export)
		decode(msg.Data(), export)

		for _, metrics := range export.Metrics {
			rm := new(v1.ResourceMetrics)
			if err := proto.Unmarshal(metrics.OTLP, rm); err != nil {
				logger.Fatal("failed to decode resource metrics protobuf", zap.Error(err))
			}
			request := v1_service.ExportMetricsServiceRequest{
				ResourceMetrics: []*v1.ResourceMetrics{rm},
			}
			requestPayload, err := proto.Marshal(&request)
			if err != nil {
				logger.Fatal("failed to marshal export metrics request", zap.Error(err))
			}

			res, err := http.Post(vmExportUrl, "application/octet-stream", bytes.NewReader(requestPayload))
			if err != nil {
				logger.Fatal("failed to POST metrics export request to victoria metrics", zap.Error(err))
			}
			if res.StatusCode != 200 {
				logger.Fatal("failed to POST metrics export request to victoria metrics, unexpected status code", zap.Int("code", res.StatusCode))
			}
		}

		msg.Ack()
	})

	if err != nil {
		logger.Error("failed to consume messages", zap.Error(err))
		return err

	}
	<-cctx.Closed()

	return nil
}

func decode(data []byte, output any) {
	if err := json.Unmarshal(data, output); err != nil {
		panic("failed to unmarshal json")
	}
}
