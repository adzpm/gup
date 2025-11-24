package netrc

import (
	"os/user"
	"path/filepath"

	logger "github.com/adzpm/glone/internal/logger"
)

// LoaderOptions holds netrc loader configuration options
type LoaderOptions struct {
	Logger    logger.Logger
	NetrcPath string
}

// LoaderOption is a function that modifies LoaderOptions
type LoaderOption func(*LoaderOptions)

// WithLogger sets the logger
func WithLogger(lgr logger.Logger) LoaderOption {
	return func(o *LoaderOptions) {
		o.Logger = lgr
	}
}

// WithNetrcPath sets a custom path to .netrc file
func WithNetrcPath(path string) LoaderOption {
	return func(o *LoaderOptions) {
		o.NetrcPath = path
	}
}

// defaultLoaderOptions returns default loader options
func defaultLoaderOptions() (*LoaderOptions, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	return &LoaderOptions{
		Logger:    nil,
		NetrcPath: filepath.Join(usr.HomeDir, ".netrc"),
	}, nil
}
