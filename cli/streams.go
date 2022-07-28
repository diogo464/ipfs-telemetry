package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var commandStreams = &cli.Command{
	Name:        "streams",
	Description: "Show available streams",
	Action:      actionStreams,
}

func actionStreams(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	streams, err := client.AvailableStreams(c.Context)
	if err != nil {
		return err
	}

	for _, stream := range streams {
		fmt.Println(stream.Name)
		fmt.Println("\tEncoding:", stream.Encoding)
		fmt.Println("\tPeriod:", stream.Period)
	}

	return nil
}
