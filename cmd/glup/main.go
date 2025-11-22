package main

import (
	"context"
	"os"

	cli "github.com/urfave/cli/v3"

	clone "github.com/adzpm/glup/internal/app/clone"
	logger "github.com/adzpm/glup/internal/logger"
)

func main() {
	// Create logger instance
	lgr := logger.New()

	app := &cli.Command{
		Name:  "glup",
		Usage: "Clones all available repositories from GitLab",
		Flags: flags(),
		Commands: []*cli.Command{
			{
				Name:      "clone",
				Usage:     "clones all available repositories",
				ArgsUsage: "[directory]",
				Flags:     clone.Flags(),
				Action:    clone.Run,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		lgr.Fatal(err)
	}
}
