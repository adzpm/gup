package netrc

import (
	"fmt"
	"os"
	"strings"

	netrc "github.com/jdx/go-netrc"

	config "github.com/adzpm/glone/internal/config"
)

// Loader handles loading credentials from .netrc
type Loader struct {
	opts *LoaderOptions
}

// NewLoader creates a new loader with options
func NewLoader(opts ...LoaderOption) (*Loader, error) {
	options, err := defaultLoaderOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to get default options: %w", err)
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Loader{opts: options}, nil
}

// LoadCredentials loads GitLab credentials from .netrc file
func (l *Loader) LoadCredentials() (*config.Config, error) {
	if _, err := os.Stat(l.opts.NetrcPath); os.IsNotExist(err) {
		return nil, nil
	}

	n, err := netrc.Parse(l.opts.NetrcPath)
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
		if l.opts.Logger != nil {
			l.opts.Logger.Warn("Found multiple GitLab entries in .netrc, will use the first one")
			for i, m := range gitlabMachines {
				l.opts.Logger.Infof("  %d. %s", i+1, m.Name)
			}
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
