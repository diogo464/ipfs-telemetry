package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	mpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

var CommandMetrics = &cli.Command{
	Name:        "metrics",
	Description: "Display latest metrics",
	Action:      actionMetrics,
}

func actionMetrics(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	rmetrics, err := client.GetMetrics(c.Context, 0)
	if err != nil {
		return err
	}
	if len(rmetrics) == 0 {
		fmt.Println("metrics len is 0 ")
		return nil
	}
	mdata := rmetrics[len(rmetrics)-1]

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

	return nil
}
