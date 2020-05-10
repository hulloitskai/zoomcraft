package types

import (
	"encoding"

	"github.com/cockroachdb/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// An ID is a unique identifier used to identify values.
type ID struct{ primitive.ObjectID }

// ZeroID is the zero value for an ID.
var ZeroID = ID{primitive.NilObjectID}

// NewID generates an unique ID string based on the current timestamp.
func NewID() ID {
	return ID{ObjectID: primitive.NewObjectID()}
}

// ParseID parses an ID from a string.
func ParseID(s string) (ID, error) {
	oid, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return ID{ObjectID: primitive.NilObjectID}, err
	}
	return ID{ObjectID: oid}, nil
}

var (
	_ encoding.TextMarshaler   = (*ID)(nil)
	_ encoding.TextUnmarshaler = (*ID)(nil)
)

// MarshalText implements encoding.TextMarshaler.
func (id ID) MarshalText() ([]byte, error) {
	return []byte(id.Hex()), nil
}

// UnmarshalText implements bson.ValueUnmarshaler.
func (id *ID) UnmarshalText(text []byte) (err error) {
	*id, err = ParseID(string(text))
	return err
}

var (
	_ bson.ValueMarshaler   = (*ID)(nil)
	_ bson.ValueUnmarshaler = (*ID)(nil)
)

// MarshalBSONValue implements bson.ValueMarshaler.
func (id ID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(id.ObjectID)
}

// UnmarshalBSONValue implements bson.ValueUnmarshaler.
func (id *ID) UnmarshalBSONValue(t bsontype.Type, val []byte) error {
	var (
		rv = bson.RawValue{Type: t, Value: val}
		ok bool
	)
	id.ObjectID, ok = rv.ObjectIDOK()
	if !ok {
		return errors.New("core: not a primitive.ObjectID")
	}
	return nil
}
