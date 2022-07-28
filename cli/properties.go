package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var commandProperties = &cli.Command{
	Name:        "properties",
	Description: "Show available properties",
	Action:      actionProperties,
}

func actionProperties(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	properties, err := client.AvailableProperties(c.Context)
	if err != nil {
		return err
	}

	for _, prop := range properties {
		fmt.Println(prop.Name)
		fmt.Println("\tEncoding:", prop.Encoding)
	}

	return nil
}
