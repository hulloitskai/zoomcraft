package minecraft

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/exthttp"
)

// Coordinates describe a point in the Minecraft world by its XYZ coordinates.
type Coordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// ParseCoordinates parses a Coordinates from a string.
func ParseCoordinates(s string) (c Coordinates, err error) {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "[](){}")

	var fields []string
	if strings.IndexByte(s, ',') >= 0 {
		fields = strings.Split(s, ",")
	} else {
		fields = strings.Fields(s)
	}

	k := len(fields)
	if k != 3 {
		return c, errors.New("minecraft: expected 3 parts")
	}

	var (
		coords Coordinates
		parts  = []*float64{&coords.X, &coords.Y, &coords.Z}
	)
	for i, f := range fields {
		if *(parts[i]), err = strconv.ParseFloat(f, 64); err != nil {
			return c, errors.Wrapf(err, "part %d", i)
		}
	}
	return coords, nil
}

var _ fmt.Stringer = (*Coordinates)(nil)

func (c Coordinates) String() string {
	return fmt.Sprintf("[%f, %f, %f]", c.X, c.Y, c.Z)
}

var (
	_ graphql.Marshaler   = (*Coordinates)(nil)
	_ graphql.Unmarshaler = (*Coordinates)(nil)
)

// MarshalGQL implements graphql.Marshaler.
func (c Coordinates) MarshalGQL(w io.Writer) { io.WriteString(w, c.String()) }

// UnmarshalGQL implements graphql.Unmarshaler.
func (c *Coordinates) UnmarshalGQL(v interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errors.WithDetail(err, "Failed to parse Coordinates.")
			err = exthttp.WrapWithHTTPCode(err, http.StatusBadRequest)
		}
	}()

	switch value := v.(type) {
	case string:
		*c, err = ParseCoordinates(value)
		return err
	case []interface{}:
		k := len(value)
		if k != 3 {
			return errors.New("minecraft: expected 3 parts")
		}

		parts := []*float64{&c.X, &c.Y, &c.Z}
		for i, v := range value {
			switch value := v.(type) {
			case float64:
				*(parts[i]) = value
			case string:
				if *(parts[i]), err = strconv.ParseFloat(value, 64); err != nil {
					return errors.Wrap(err, "parse float")
				}
			case json.Number:
				if *(parts[i]), err = value.Float64(); err != nil {
					return errors.Wrap(err, "parse float")
				}
			}
		}
		return nil

	default:
		return errors.Newf("minecraft: unsupported field type %T", value)
	}
}
