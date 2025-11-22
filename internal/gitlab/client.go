package gitlab

import (
	"fmt"

	"github.com/adzpm/glup/internal/config"
	"github.com/charmbracelet/log"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Client wraps GitLab API client
type Client struct {
	*gitlab.Client
	logger *log.Logger
}

// NewClient creates a new GitLab client and authenticates
func NewClient(cfg *config.Config, logger *log.Logger) (*Client, error) {
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

	logger.Infof("Authenticated as: %s (%s)", user.Username, user.Email)
	return &Client{
		Client: client,
		logger: logger,
	}, nil
}
