package netrc

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/adzpm/glup/internal/config"
	"github.com/charmbracelet/log"
	"github.com/jdx/go-netrc"
)

// LoadCredentials loads GitLab credentials from .netrc file
func LoadCredentials(logger *log.Logger) (*config.Config, error) {
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
		logger.Warn("Found multiple GitLab entries in .netrc, will use the first one")
		for i, m := range gitlabMachines {
			logger.Infof("  %d. %s", i+1, m.Name)
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

	return &config.Config{
		GitLabHost:  host,
		GitLabUser:  login,
		GitLabToken: password,
	}, nil
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
