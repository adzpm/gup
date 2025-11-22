package main

import (
	"github.com/urfave/cli/v3"
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
