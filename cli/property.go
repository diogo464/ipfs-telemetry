package cli

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

var commandProperty = &cli.Command{
	Name:        "property",
	Description: "retreive a property from a node",
	Action:      actionProperty,
	Flags:       []cli.Flag{},
}

func actionProperty(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	data, err := client.Property(c.Context, c.Args().First())
	if err != nil {
		return err
	}
	fmt.Println("Data:", string(data))

	var obj map[string]interface{} = make(map[string]interface{})
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}

	fmt.Println(obj)

	return nil
}
