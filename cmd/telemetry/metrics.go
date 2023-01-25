package main

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

var FLAG_METRICS_RAW = &cli.BoolFlag{
	Name:  "raw",
	Usage: "Display raw metrics",
}

var CommandMetrics = &cli.Command{
	Name:        "metrics",
	Description: "Display latest metrics",
	Action:      actionMetrics,
	Flags: []cli.Flag{
		FLAG_METRICS_RAW,
	},
}

func actionMetrics(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	cmetrics, err := client.GetMetrics(c.Context)
	if err != nil {
		return err
	}
	metrics := cmetrics.OTLP

	if c.Bool(FLAG_METRICS_RAW.Name) {
		m, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(m))
	} else {
		if len(metrics) == 0 {
			fmt.Println("metrics len is 0 ")
			return nil
		}
		mdata := metrics[len(metrics)-1]

		for _, scope := range mdata.ScopeMetrics {
			for _, metric := range scope.Metrics {
				fmt.Println(scope.Scope.Name + "/" + metric.Name)
				switch v := metric.GetData().(type) {
				case *mpb.Metric_Gauge:
					if len(v.Gauge.DataPoints) > 0 {
						fmt.Println(v.Gauge.DataPoints[0].GetValue())
					} else {
						fmt.Println(0)
					}
				case *mpb.Metric_Sum:
					for _, dp := range v.Sum.DataPoints {
						fmt.Println("\t", dp.Attributes, " = ", dp.GetValue())
					}
				case *mpb.Metric_Summary:
					fmt.Println("summary")
				case *mpb.Metric_Histogram:
					fmt.Println("histogram")
				case *mpb.Metric_ExponentialHistogram:
					fmt.Println("exp histogram")
				default:
					fmt.Println("unknown")
				}
			}
		}
	}

	return nil
}
