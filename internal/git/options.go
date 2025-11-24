package git

import (
	"io"
	"os"

	logger "github.com/adzpm/glone/internal/logger"
)

// ClonerOptions holds cloner configuration options
type ClonerOptions struct {
	Logger      logger.Logger
	ProgressOut io.Writer
}

// ClonerOption is a function that modifies ClonerOptions
type ClonerOption func(*ClonerOptions)

// WithLogger sets the logger
func WithLogger(lgr logger.Logger) ClonerOption {
	return func(o *ClonerOptions) {
		o.Logger = lgr
	}
}

// WithProgressOutput sets the progress output writer
func WithProgressOutput(w io.Writer) ClonerOption {
	return func(o *ClonerOptions) {
		o.ProgressOut = w
	}
}

// defaultClonerOptions returns default cloner options
func defaultClonerOptions() *ClonerOptions {
	return &ClonerOptions{
		Logger:      nil,
		ProgressOut: os.Stdout,
	}
}
