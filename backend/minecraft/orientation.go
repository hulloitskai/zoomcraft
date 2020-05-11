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

// Orientation describe an orientation in 3D space.
type Orientation struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// ParseOrientation parses an Orientation from a string.
func ParseOrientation(s string) (o Orientation, err error) {
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
		return o, errors.New("minecraft: expected 3 parts")
	}

	var (
		orient Orientation
		parts  = []*float32{&orient.X, &orient.Y}
	)
	for i, f := range fields {
		f64, err := strconv.ParseFloat(f, 32)
		if err != nil {
			return o, errors.Wrapf(err, "part %d", i)
		}
		*(parts[i]) = float32(f64)
	}
	return orient, nil
}

var _ fmt.Stringer = (*Orientation)(nil)

func (c Orientation) String() string {
	return fmt.Sprintf("[%f, %f]", c.X, c.Y)
}

var (
	_ graphql.Marshaler   = (*Orientation)(nil)
	_ graphql.Unmarshaler = (*Orientation)(nil)
)

// MarshalGQL implements graphql.Marshaler.
func (c Orientation) MarshalGQL(w io.Writer) { io.WriteString(w, c.String()) }

// UnmarshalGQL implements graphql.Unmarshaler.
func (c *Orientation) UnmarshalGQL(v interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errors.WithDetail(err, "Failed to parse Orientation.")
			err = exthttp.WrapWithHTTPCode(err, http.StatusBadRequest)
		}
	}()

	switch value := v.(type) {
	case string:
		*c, err = ParseOrientation(value)
		return err
	case []interface{}:
		k := len(value)
		if k != 3 {
			return errors.New("minecraft: expected 3 parts")
		}

		parts := []*float32{&c.X, &c.Y}
		for i, v := range value {
			switch value := v.(type) {
			case float64:
				*(parts[i]) = float32(value)
			case string:
				f64, err := strconv.ParseFloat(value, 32)
				if err != nil {
					return errors.Wrap(err, "parse float")
				}
				*(parts[i]) = float32(f64)
			case json.Number:
				f64, err := value.Float64()
				if err != nil {
					return errors.Wrap(err, "parse float")
				}
				*(parts[i]) = float32(f64)
			}
		}
		return nil

	default:
		return errors.Newf("minecraft: unsupported field type %T", value)
	}
}
