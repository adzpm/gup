package clone

import (
	"context"
	"fmt"
	"os"

	cli "github.com/urfave/cli/v3"

	config "github.com/adzpm/glup/internal/config"
	git "github.com/adzpm/glup/internal/git"
	gitlab "github.com/adzpm/glup/internal/gitlab"
	logger "github.com/adzpm/glup/internal/logger"
	netrc "github.com/adzpm/glup/internal/netrc"
)

func Run(ctx context.Context, cmd *cli.Command) error {
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
	netrcLoader, err := netrc.NewLoader(netrc.WithLogger(lgr))
	if err != nil {
		return fmt.Errorf("error creating netrc loader: %w", err)
	}
	netrcCfg, err := netrcLoader.LoadCredentials()
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
	client, err := gitlab.NewClient(cfg, gitlab.WithLogger(lgr))
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

	// Create cloner
	cloner := git.NewCloner(git.WithLogger(lgr))

	for _, project := range projects {
		skipped, err := cloner.CloneProject(project, cfg.TargetDir, cfg.GitLabToken)
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
