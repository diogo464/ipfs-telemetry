package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
	v1_service "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	v1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var FLAG_VM_URL = &cli.StringFlag{
	Name:    "vm-url",
	Usage:   "victoria metrics url",
	Value:   "http://localhost:8428",
	EnvVars: []string{"VM_URL"},
}

var CommandMetricsToVm = &cli.Command{
	Name:        "metrics-to-vm",
	Description: "Export metrics to VictoriaMetrics",
	Action:      actionMetricsToVm,
	Flags: []cli.Flag{
		FLAG_VM_URL,
	},
}

func actionMetricsToVm(c *cli.Context) error {
	logger, _ := zap.NewProduction()
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	messages, err := client.GetMetricsMessages(c.Context)
	if err != nil {
		return err
	}

	vmExportUrl := fmt.Sprintf("%s/opentelemetry/v1/metrics", c.String(FLAG_VM_URL.Name))
	for _, msg := range messages {
		fmt.Printf("exporting %v %v bytes\n", msg.Timestamp.Format("2006-01-02 15:04:05"), len(msg.Value))
		m := &v1.ResourceMetrics{}
		err = proto.Unmarshal(msg.Value, m)
		if err != nil {
			return err
		}

		request := v1_service.ExportMetricsServiceRequest{
			ResourceMetrics: []*v1.ResourceMetrics{m},
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

	return nil
}
