package clone

import (
	cli "github.com/urfave/cli/v3"
)

// Flags returns flags for the clone command
func Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "group",
			Usage: "clone repositories only from specified group",
		},
	}
}
