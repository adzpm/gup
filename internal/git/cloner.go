package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
)

// Cloner handles git cloning operations
type Cloner struct {
	opts *ClonerOptions
}

// NewCloner creates a new cloner with options
func NewCloner(opts ...ClonerOption) *Cloner {
	options := defaultClonerOptions()
	for _, opt := range opts {
		opt(options)
	}
	return &Cloner{opts: options}
}

// CloneProject clones a project to the target directory
func (c *Cloner) CloneProject(project *Project, targetDir string, token string) (bool, error) {
	// Use PathWithNamespace to preserve directory structure
	projectPath := filepath.Join(targetDir, project.PathWithNamespace)

	// Check if directory already exists and if it's a git repository
	if info, err := os.Stat(projectPath); err == nil {
		if info.IsDir() {
			// Check if this is a git repository
			if _, err := git.PlainOpen(projectPath); err == nil {
				if c.opts.Logger != nil {
					c.opts.Logger.Warnf("Project %s already exists in %s, skipping", project.Name, projectPath)
				}
				return true, nil // true means the project was skipped
			}
			// If directory exists but is not a git repository, remove it
			if c.opts.Logger != nil {
				c.opts.Logger.Warnf("Directory %s exists but is not a git repository, removing", projectPath)
			}
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

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(projectPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}

	// Clone repository (git.PlainClone creates the final directory itself)
	if c.opts.Logger != nil {
		c.opts.Logger.Infof("Cloning %s to %s", project.Name, projectPath)
	}
	_, err := git.PlainClone(projectPath, false, &git.CloneOptions{
		URL:      cloneURL,
		Progress: c.opts.ProgressOut,
	})

	if err != nil {
		// Remove directory on error
		os.RemoveAll(projectPath)
		return false, fmt.Errorf("error cloning %s: %w", project.Name, err)
	}

	if c.opts.Logger != nil {
		c.opts.Logger.Infof("Successfully cloned: %s", project.Name)
	}
	return false, nil // false means the project was successfully cloned
}
