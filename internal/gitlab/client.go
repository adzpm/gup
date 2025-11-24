package gitlab

import (
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	config "github.com/adzpm/glone/internal/config"
)

// Client wraps GitLab API client
type Client struct {
	*gitlab.Client
	logger *ClientOptions
}

// NewClient creates a new GitLab client and authenticates
func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	baseURL := options.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s", cfg.GitLabHost)
	}

	client, err := gitlab.NewClient(cfg.GitLabToken, gitlab.WithBaseURL(baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Check connection if not skipped
	if !options.SkipAuth {
		user, _, err := client.Users.CurrentUser()
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate with GitLab: %w", err)
		}

		if options.Logger != nil {
			options.Logger.Infof("Authenticated as: %s (%s)", user.Username, user.Email)
		}
	}

	return &Client{
		Client: client,
		logger: options,
	}, nil
}
