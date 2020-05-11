package logutil

import (
	"fmt"

	"github.com/go-kit/kit/log"
)

const messageKey = "msg"

// Log writes a log message to logger.
func Log(logger log.Logger, format string, args ...interface{}) error {
	return logger.Log(messageKey, fmt.Sprintf(format, args...))
}

// Trace writes a log message describing a method trace.
func Trace(logger log.Logger, method string, err error) error {
	logger = WithError(logger, err)
	return logger.Log(messageKey, fmt.Sprintf("Executed method: %s", method))
}
