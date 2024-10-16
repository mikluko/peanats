package jetbucket

import (
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

type entry interface {
	jetstream.KeyValueEntry
}

type marshaler interface {
	Marshal() ([]byte, error)
}

func marshal(x any) ([]byte, error) {
	m, ok := x.(marshaler)
	if !ok {
		return nil, fmt.Errorf("%T doesn't implement marshaler", x)
	}
	return m.Marshal()
}

type unmarshaler interface {
	Unmarshal([]byte) error
}

func unmarshal(x any, data []byte) error {
	u, ok := x.(unmarshaler)
	if !ok {
		return fmt.Errorf("%T doesn't implement unmarshaler", x)
	}
	return u.Unmarshal(data)
}
