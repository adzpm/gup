package logger

import (
	"os"
	"time"

	log "github.com/charmbracelet/log"
)

// Adapter wraps charmbracelet/log.Logger to implement Logger interface
type Adapter struct {
	*log.Logger
}

// NewAdapter creates a new adapter from charmbracelet/log.Logger
func NewAdapter(l *log.Logger) *Adapter {
	return &Adapter{Logger: l}
}

// New creates and initializes a new logger instance with options
func New(opts ...Option) Logger {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	l := log.NewWithOptions(os.Stdout, log.Options{
		Prefix:          "glone",
		TimeFormat:      time.Kitchen,
		Level:           log.DebugLevel,
		ReportTimestamp: true,
	})

	return NewAdapter(l)
}
