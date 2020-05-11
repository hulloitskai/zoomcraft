package types

import (
	"time"

	"github.com/fatih/structs"
)

type (
	// Any is an alias for the empty interface.
	Any = interface{}

	// Empty is an alias for the empty struct.
	Empty = struct{}
)

// A Zeroer can determine if it a zero-value.
//
// One such common instance of a Zeroer is a time.Time.
type Zeroer interface {
	IsZero() bool
}

var _ Zeroer = (*time.Time)(nil)

// IsZero returns true if v is a zero-value.
func IsZero(v interface{}) bool {
	if z, ok := v.(Zeroer); ok {
		return z.IsZero()
	}
	return structs.IsZero(v)
}
