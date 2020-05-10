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
