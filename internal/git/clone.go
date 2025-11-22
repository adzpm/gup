package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/go-git/go-git/v5"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// CloneResult represents the result of a clone operation
type CloneResult struct {
	Skipped bool
	Error   error
}

// CloneProject clones a GitLab project to the target directory
func CloneProject(project *gitlab.Project, targetDir string, token string, logger *log.Logger) (bool, error) {
	// Use PathWithNamespace to preserve directory structure (e.g., helios/tests/atlassian/jira)
	projectPath := filepath.Join(targetDir, project.PathWithNamespace)

	// Check if directory already exists and if it's a git repository
	if info, err := os.Stat(projectPath); err == nil {
		if info.IsDir() {
			// Check if this is a git repository
			if _, err := git.PlainOpen(projectPath); err == nil {
				logger.Warnf("Project %s already exists in %s, skipping", project.Name, projectPath)
				return true, nil // true means the project was skipped
			}
			// If directory exists but is not a git repository, remove it
			logger.Warnf("Directory %s exists but is not a git repository, removing", projectPath)
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
	logger.Infof("Cloning %s to %s", project.Name, projectPath)
	_, err := git.PlainClone(projectPath, false, &git.CloneOptions{
		URL:      cloneURL,
		Progress: os.Stdout,
	})

	if err != nil {
		// Remove directory on error
		os.RemoveAll(projectPath)
		return false, fmt.Errorf("error cloning %s: %w", project.Name, err)
	}

	logger.Infof("Successfully cloned: %s", project.Name)
	return false, nil // false means the project was successfully cloned
}
