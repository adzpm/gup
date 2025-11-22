package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	"github.com/jdx/go-netrc"
	"github.com/urfave/cli/v3"
	gitlab "github.com/xanzy/go-gitlab"
)

type Config struct {
	GitLabHost  string
	GitLabUser  string
	GitLabToken string
	Group       string
	TargetDir   string
}

func findGitLabCredentials(n *netrc.Netrc) ([]*netrc.Machine, error) {
	var gitlabMachines []*netrc.Machine

	for _, machine := range n.Machines() {
		machineName := strings.ToLower(machine.Name)
		if strings.Contains(machineName, "gitlab") {
			gitlabMachines = append(gitlabMachines, machine)
		}
	}

	return gitlabMachines, nil
}

func loadNetrcCredentials() (*Config, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	netrcPath := filepath.Join(usr.HomeDir, ".netrc")
	if _, err := os.Stat(netrcPath); os.IsNotExist(err) {
		return nil, nil
	}

	n, err := netrc.Parse(netrcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .netrc: %w", err)
	}

	gitlabMachines, err := findGitLabCredentials(n)
	if err != nil {
		return nil, fmt.Errorf("error searching for GitLab credentials: %w", err)
	}

	if len(gitlabMachines) == 0 {
		return nil, nil
	}

	if len(gitlabMachines) > 1 {
		log.Warn("Found multiple GitLab entries in .netrc, will use the first one")
		for i, m := range gitlabMachines {
			log.Infof("  %d. %s", i+1, m.Name)
		}
	}

	machine := gitlabMachines[0]
	login := machine.Get("login")
	password := machine.Get("password")

	if login == "" || password == "" {
		return nil, fmt.Errorf("incomplete credentials for %s in .netrc", machine.Name)
	}

	// Determine host from machine name or use default
	host := machine.Name
	if host == "" {
		host = "gitlab.com"
	}

	return &Config{
		GitLabHost:  host,
		GitLabUser:  login,
		GitLabToken: password,
	}, nil
}

func createGitLabClient(cfg *Config) (*gitlab.Client, error) {
	baseURL := fmt.Sprintf("https://%s", cfg.GitLabHost)
	client, err := gitlab.NewClient(cfg.GitLabToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Check connection
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with GitLab: %w", err)
	}

	log.Infof("Authenticated as: %s (%s)", user.Username, user.Email)
	return client, nil
}

func getAllProjects(client *gitlab.Client, groupName string) ([]*gitlab.Project, error) {
	var allProjects []*gitlab.Project

	// If group is specified, get projects from group
	if groupName != "" {
		groupOpt := &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				PerPage: 100,
				Page:    1,
			},
			Archived: gitlab.Bool(true),
		}

		for {
			projects, resp, err := client.Groups.ListGroupProjects(groupName, groupOpt)
			if err != nil {
				return nil, fmt.Errorf("error getting projects for group %s: %w", groupName, err)
			}

			allProjects = append(allProjects, projects...)

			if resp.NextPage == 0 {
				break
			}

			groupOpt.Page = resp.NextPage
		}

		return allProjects, nil
	}

	// Otherwise get all available projects
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
			Page:    1,
		},
		Archived: gitlab.Bool(true), // Include archived
		Simple:   gitlab.Bool(false),
	}

	for {
		projects, resp, err := client.Projects.ListProjects(opt)
		if err != nil {
			return nil, fmt.Errorf("error getting project list: %w", err)
		}

		allProjects = append(allProjects, projects...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return allProjects, nil
}

func cloneProject(project *gitlab.Project, targetDir string, token string) (bool, error) {
	projectPath := filepath.Join(targetDir, project.Path)

	// Check if directory already exists and if it's a git repository
	if info, err := os.Stat(projectPath); err == nil {
		if info.IsDir() {
			// Check if this is a git repository
			if _, err := git.PlainOpen(projectPath); err == nil {
				log.Warnf("Project %s already exists in %s, skipping", project.Name, projectPath)
				return true, nil // true means the project was skipped
			}
			// If directory exists but is not a git repository, remove it
			log.Warnf("Directory %s exists but is not a git repository, removing", projectPath)
			os.RemoveAll(projectPath)
		}
	}

	// Form URL with token for cloning
	cloneURL := project.HTTPURLToRepo
	if token != "" && strings.HasPrefix(cloneURL, "http") {
		// Add token to URL for authentication
		if strings.HasPrefix(cloneURL, "https://") {
			cloneURL = strings.Replace(cloneURL, "https://", fmt.Sprintf("https://oauth2:%s@", token), 1)
		} else if strings.HasPrefix(cloneURL, "http://") {
			cloneURL = strings.Replace(cloneURL, "http://", fmt.Sprintf("http://oauth2:%s@", token), 1)
		}
	}

	// Clone repository (git.PlainClone creates directory itself)
	log.Infof("Cloning %s to %s", project.Name, projectPath)
	_, err := git.PlainClone(projectPath, false, &git.CloneOptions{
		URL:      cloneURL,
		Progress: os.Stdout,
	})

	if err != nil {
		// Remove directory on error
		os.RemoveAll(projectPath)
		return false, fmt.Errorf("error cloning %s: %w", project.Name, err)
	}

	log.Infof("Successfully cloned: %s", project.Name)
	return false, nil // false means the project was successfully cloned
}

func cloneCommand(ctx context.Context, cmd *cli.Command) error {
	cfg := &Config{
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
	if cfg.GitLabHost == "" || cfg.GitLabUser == "" || cfg.GitLabToken == "" {
		netrcCfg, err := loadNetrcCredentials()
		if err != nil {
			return fmt.Errorf("error loading .netrc: %w", err)
		}

		if netrcCfg != nil {
			if cfg.GitLabHost == "" {
				cfg.GitLabHost = netrcCfg.GitLabHost
			}
			if cfg.GitLabUser == "" {
				cfg.GitLabUser = netrcCfg.GitLabUser
			}
			if cfg.GitLabToken == "" {
				cfg.GitLabToken = netrcCfg.GitLabToken
			}
			log.Info("Using credentials from .netrc")
		}
	}

	// Check that all required data is specified
	if cfg.GitLabHost == "" || cfg.GitLabUser == "" || cfg.GitLabToken == "" {
		return fmt.Errorf("must specify --gitlab-host, --gitlab-user and --gitlab-token or configure .netrc")
	}

	// Create GitLab client
	client, err := createGitLabClient(cfg)
	if err != nil {
		return err
	}

	// Get project list
	log.Info("Getting project list...")
	projects, err := getAllProjects(client, cfg.Group)
	if err != nil {
		return err
	}

	log.Infof("Found projects: %d", len(projects))

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(cfg.TargetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Clone each project
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, project := range projects {
		skipped, err := cloneProject(project, cfg.TargetDir, cfg.GitLabToken)
		if err != nil {
			log.Errorf("Error cloning %s: %v", project.Name, err)
			errorCount++
		} else if skipped {
			skipCount++
		} else {
			successCount++
		}
	}

	log.Infof("Completed. Success: %d, Skipped: %d, Errors: %d", successCount, skipCount, errorCount)

	return nil
}

func main() {
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
		log.Fatal(err)
	}
}
