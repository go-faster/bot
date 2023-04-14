package action

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/go-faster/errors"
)

//go:generate go run github.com/dmarkham/enumer -type Entity -output entity_str_gen.go

type Entity uint8

const (
	Unknown Entity = iota
	PullRequest
	Issue
)

//go:generate go run github.com/dmarkham/enumer -type Type -output type_str_gen.go

type Type uint8

const (
	UnknownType Type = iota
	Merge
	Close
)

type Action struct {
	Entity       Entity
	Type         Type
	ID           int
	RepositoryID int64
}

func (a Action) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s(%s=%d)", a.Type, a.Entity, a.ID))
	if a.RepositoryID != 0 {
		b.WriteString(fmt.Sprintf(" r=%d", a.RepositoryID))
	}
	return b.String()
}

func (a Action) MarshalText() ([]byte, error) {
	data, err := a.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}
	return []byte(base64.RawStdEncoding.EncodeToString(data)), nil
}

func (a *Action) UnmarshalText(data []byte) error {
	b, err := base64.RawStdEncoding.DecodeString(string(data))
	if err != nil {
		return errors.Wrap(err, "decode")
	}
	return a.UnmarshalBinary(b)
}

const binarySize = 18

func (a *Action) UnmarshalBinary(data []byte) error {
	// Inverse of MarshalBinary.
	if len(data) != binarySize {
		return errors.New("invalid data length")
	}
	a.Entity = Entity(data[0])
	a.Type = Type(data[1])
	a.ID = int(binary.BigEndian.Uint64(data[2:10]))
	a.RepositoryID = int64(binary.BigEndian.Uint64(data[10:18]))
	return nil
}

func (a Action) MarshalBinary() ([]byte, error) {
	var out [binarySize]byte
	out[0] = byte(a.Entity)
	out[1] = byte(a.Type)
	binary.BigEndian.PutUint64(out[2:10], uint64(a.ID))
	binary.BigEndian.PutUint64(out[10:18], uint64(a.RepositoryID))
	return out[:], nil
}

func Marshal(a Action) []byte {
	data, _ := a.MarshalBinary()
	return data
}
