package logging

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// NewLogger creates a logger with the specified minimum level.
// If verbose is true, logs at debug level; otherwise info level.
func NewLogger(verbose bool) log.Logger {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	if verbose {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	return logger
}
