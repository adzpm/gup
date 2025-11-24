package main

import (
	"context"
	"os"

	cli "github.com/urfave/cli/v3"

	clone "github.com/adzpm/glone/internal/app/clone"
	logger "github.com/adzpm/glone/internal/logger"
)

// flags returns global flags for the application
func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "gitlab-host",
			Usage:   "GitLab host (e.g., gitlab.com)",
			Sources: cli.EnvVars("GITLAB_HOST"),
		},
		&cli.StringFlag{
			Name:    "gitlab-user",
			Usage:   "GitLab user",
			Sources: cli.EnvVars("GITLAB_USER"),
		},
		&cli.StringFlag{
			Name:    "gitlab-token",
			Usage:   "GitLab access token",
			Sources: cli.EnvVars("GITLAB_TOKEN"),
		},
	}
}

func main() {
	// Create logger instance
	lgr := logger.New()

	app := &cli.Command{
		Name:  "glone",
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
