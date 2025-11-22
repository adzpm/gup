package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

// New creates and initializes a new logger instance
func New() *log.Logger {
	l := log.New(os.Stderr)
	l.SetLevel(log.InfoLevel)
	l.SetReportTimestamp(true)
	l.SetReportCaller(false)
	return l
}
