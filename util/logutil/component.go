package logutil

import (
	"strings"

	"github.com/go-kit/kit/log"
)

const componentKey = "component"

// WithComponent adds a component field to a log.Logger.
func WithComponent(logger log.Logger, name string) log.Logger {
	return log.LoggerFunc(func(keyvals ...interface{}) error {
		if len(keyvals)%2 != 0 {
			keyvals = append(keyvals, log.ErrMissingValue)
		}
		var b strings.Builder
		b.WriteString(name)

		kvs := make([]interface{}, 0, len(keyvals)+2)
		for i := 0; i < len(keyvals); i += 2 {
			if keyvals[i] == componentKey {
				b.WriteByte('/')
				b.WriteString(keyvals[i+1].(string))
			} else {
				kvs = append(kvs, keyvals[i], keyvals[i+1])
			}
		}
		kvs = append(kvs, componentKey, b.String())
		return logger.Log(kvs...)
	})
}
