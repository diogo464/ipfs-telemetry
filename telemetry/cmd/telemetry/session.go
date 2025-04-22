package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var CommandSession = &cli.Command{
	Name:   "session",
	Action: actionSessionInfo,
}

func actionSessionInfo(c *cli.Context) error {
	client, err := clientFromContext(c)
	if err != nil {
		return err
	}
	defer client.Close()

	sess, err := client.GetSession(c.Context)
	if err != nil {
		return err
	}
	fmt.Println(sess.String())
	return nil
}
