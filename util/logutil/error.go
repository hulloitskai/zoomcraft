package logutil

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const errorKey = "err"

// WithError adds an error field to a log.Logger, and sets it to log at the
// error level.
func WithError(logger log.Logger, err error) log.Logger {
	if err == nil {
		return logger
	}
	logger = level.NewInjector(logger, level.ErrorValue())
	return log.With(logger, errorKey, err)
}
