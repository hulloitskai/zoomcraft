package build

import (
	"time"
)

var (
	// version is the build version, and is set at compile-time with the following
	// linker flag:
	//   -X go.stevenxie.me/covidcraft/internal/build.version=$VERSION
	version string

	// timestamp is the build timestamp, and is set at compile-time with the
	// following linker flag
	//   -X go.stevenxie.me/covidcraft/internal/build.timestamp=$TIMESTAMP
	timestamp string
)

// Info contains build information.
type Info struct {
	Version   string    `json:"version,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ReadInfo reads build info set at compile-time.
func ReadInfo() *Info {
	info := &Info{Version: version}
	if timestamp != "" {
		var err error
		if info.Timestamp, err = time.Parse(time.RFC3339, timestamp); err != nil {
			panic(err)
		}
	}
	return info
}
