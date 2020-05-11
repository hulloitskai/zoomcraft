package logutil

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-logfmt/logfmt"
)

type encoder struct {
	*logfmt.Encoder
	buf bytes.Buffer
}

func (enc *encoder) Reset() {
	enc.Encoder.Reset()
	enc.buf.Reset()
}

var encoderPool = sync.Pool{
	New: func() interface{} {
		var enc encoder
		enc.Encoder = logfmt.NewEncoder(&enc.buf)
		return &enc
	},
}

type encodedLogger struct {
	w io.Writer
}

// NewLogger returns a Logger that writes logs to w.
func NewLogger(w io.Writer) log.Logger {
	return &encodedLogger{w}
}

func (l encodedLogger) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 == 1 {
		keyvals = append(keyvals, nil)
	}

	enc := encoderPool.Get().(*encoder)
	enc.Reset()
	defer encoderPool.Put(enc)

	var (
		message   string
		component string
	)
	for i := 0; i < len(keyvals); i += 2 {
		k, v := keyvals[i], keyvals[i+1]

		// Account for custom keys.
		if k == componentKey {
			component = fmt.Sprint(v)
			continue
		}
		if k == messageKey {
			message = fmt.Sprint(v)
		}

		err := enc.EncodeKeyval(k, v)
		if err == logfmt.ErrUnsupportedKeyType {
			continue
		}
		if _, ok := err.(*logfmt.MarshalerError); ok || err == logfmt.ErrUnsupportedValueType {
			v = err
			err = enc.EncodeKeyval(k, v)
		}
		if err != nil {
			return err
		}
	}
	if err := enc.EncodeKeyvals(keyvals...); err != nil {
		return err
	}

	// Add newline to the end of the buffer
	if err := enc.EndRecord(); err != nil {
		return err
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[%s] %s ", component, message)
	enc.buf.WriteTo(&buf)

	if _, err := l.w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
