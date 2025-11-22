package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adzpm/glup/internal/config"
	"github.com/adzpm/glup/internal/git"
	"github.com/adzpm/glup/internal/gitlab"
	"github.com/adzpm/glup/internal/logger"
	"github.com/adzpm/glup/internal/netrc"
	"github.com/urfave/cli/v3"
)

func cloneCommand(ctx context.Context, cmd *cli.Command) error {
	// Create logger instance
	lgr := logger.New()
	cfg := &config.Config{
		GitLabHost:  cmd.String("gitlab-host"),
		GitLabUser:  cmd.String("gitlab-user"),
		GitLabToken: cmd.String("gitlab-token"),
		Group:       cmd.String("group"),
		TargetDir:   cmd.Args().First(),
	}

	// If TargetDir is not specified, use current directory
	if cfg.TargetDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		cfg.TargetDir = wd
	}

	// Load credentials from .netrc if not specified via flags
	netrcCfg, err := netrc.LoadCredentials(lgr)
	if err != nil {
		return fmt.Errorf("error loading .netrc: %w", err)
	}

	if netrcCfg != nil {
		cfg.Merge(netrcCfg)
		lgr.Info("Using credentials from .netrc")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w (must specify --gitlab-host, --gitlab-user and --gitlab-token or configure .netrc)", err)
	}

	// Create GitLab client
	client, err := gitlab.NewClient(cfg, lgr)
	if err != nil {
		return err
	}

	// Get project list
	lgr.Info("Getting project list...")
	projects, err := client.GetAllProjects(cfg.Group)
	if err != nil {
		return err
	}

	lgr.Infof("Found projects: %d", len(projects))

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(cfg.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Clone each project
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, project := range projects {
		skipped, err := git.CloneProject(project, cfg.TargetDir, cfg.GitLabToken, lgr)
		if err != nil {
			lgr.Errorf("Error cloning %s: %v", project.Name, err)
			errorCount++
		} else if skipped {
			skipCount++
		} else {
			successCount++
		}
	}

	lgr.Infof("Completed. Success: %d, Skipped: %d, Errors: %d", successCount, skipCount, errorCount)

	return nil
}

func main() {
	// Create logger instance
	lgr := logger.New()

	app := &cli.Command{
		Name:  "glup",
		Usage: "Clones all available repositories from GitLab",
		Flags: []cli.Flag{
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
		},
		Commands: []*cli.Command{
			{
				Name:      "clone",
				Usage:     "Clones all available repositories",
				ArgsUsage: "[directory]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "group",
						Usage: "Clone repositories only from specified group",
					},
				},
				Action: cloneCommand,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		lgr.Fatal(err)
	}
}
