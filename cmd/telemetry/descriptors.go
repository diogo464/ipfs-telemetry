package main

import "github.com/urfave/cli/v2"

var CommandDescriptors = &cli.Command{
	Name:        "descriptors",
	Description: "List all descriptors",
	Action:      actionDescriptors,
}

func actionDescriptors(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	descriptors := make([]interface{}, 0)
	if md, err := client.GetMetricDescriptors(c.Context); err == nil {
		for _, d := range md {
			descriptors = append(descriptors, createDescriptor("metric", d))
		}
	} else {
		return err
	}

	if pd, err := client.GetPropertyDescriptors(c.Context); err == nil {
		for _, d := range pd {
			descriptors = append(descriptors, createDescriptor("property", d))
		}
	} else {
		return err
	}

	if cd, err := client.GetCaptureDescriptors(c.Context); err == nil {
		for _, d := range cd {
			descriptors = append(descriptors, createDescriptor("capture", d))
		}
	} else {
		return err
	}

	if ed, err := client.GetEventDescriptors(c.Context); err == nil {
		for _, d := range ed {
			descriptors = append(descriptors, createDescriptor("event", d))
		}
	} else {
		return err
	}

	printAsJson(descriptors)

	return nil
}

func createDescriptor(kind string, desc interface{}) interface{} {
	return struct {
		Kind       string
		Descriptor interface{}
	}{
		Kind:       kind,
		Descriptor: desc,
	}
}
