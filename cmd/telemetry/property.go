package main

import (
	"fmt"

	"github.com/diogo464/telemetry"
	"github.com/urfave/cli/v2"
)

var CommandProperty = &cli.Command{
	Name:        "property",
	Description: "Show property value",
	Action:      actionProperty,
}

func actionProperty(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	properties, err := client.GetAvailableProperties(c.Context)
	if err != nil {
		return err
	}

	for _, prop := range properties {
		fmt.Print(prop.Name, "=")
		switch prop.Encoding {
		case telemetry.EncodingInt64:
			if v, err := telemetry.PropertyDecoded(c.Context, client, prop.Name, telemetry.Int64PropertyDecoder); err == nil {
				fmt.Println(v)

			} else {
				return err
			}
			break
		case telemetry.EncodingJson:
		case telemetry.EncodingString:
			if v, err := telemetry.PropertyDecoded(c.Context, client, prop.Name, telemetry.StrPropertyDecoder); err == nil {
				fmt.Println(v)

			} else {
				return err
			}
			break
		default:
			fmt.Println("Cant display property with encoding: ", telemetry.ReadableEncoding(prop.Encoding))
		}
	}

	return nil
}
