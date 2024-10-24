package jetbucket

import (
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type entry interface {
	jetstream.KeyValueEntry
}

type Entry[T any] interface {
	Bucket() string
	Key() string
	Value() *T
	Revision() uint64
	Created() time.Time
	Delta() uint64
	Operation() jetstream.KeyValueOp
}

var _ Entry[any] = (*entryImpl[any])(nil)

func EntryFromRaw[T any](raw jetstream.KeyValueEntry, opts ...EntryOption) (Entry[T], error) {
	p := entryParams{}
	for _, opt := range opts {
		opt(&p)
	}
	v := entryImpl[T]{
		entry: raw,
		key:   raw.Key(),
	}
	if p.prefix != "" {
		v.key = strings.TrimPrefix(v.key, p.prefix+".")
	}
	if raw.Operation() != jetstream.KeyValuePut {
		return v, nil
	}
	v.value = new(T)
	return v, unmarshal(any(v.value), raw.Value())
}

type entryParams struct {
	prefix string
}

type EntryOption func(params *entryParams)

func EntryPrefix(prefix string) EntryOption {
	return func(params *entryParams) {
		params.prefix = prefix
	}
}

type entryImpl[T any] struct {
	entry
	key   string
	value *T
}

func (e entryImpl[T]) Key() string {
	return e.key
}

func (e entryImpl[T]) RawValue() []byte {
	return e.entry.Value()
}

func (e entryImpl[T]) Value() *T {
	return e.value
}

type marshaler interface {
	Marshal() ([]byte, error)
}

var ErrMarshal = fmt.Errorf("marshal error")

func marshal(x any) ([]byte, error) {
	m, ok := x.(marshaler)
	if !ok {
		return nil, fmt.Errorf("%w: %T doesn't implement marshaler", ErrMarshal, x)
	}
	p, err := m.Marshal()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMarshal, err)
	}
	return p, nil
}

var ErrUnmarshal = fmt.Errorf("unmarshal error")

type unmarshaler interface {
	Unmarshal([]byte) error
}

func unmarshal(x any, data []byte) error {
	u, ok := x.(unmarshaler)
	if !ok {
		return fmt.Errorf("%w: %T doesn't implement unmarshaler", ErrUnmarshal, x)
	}
	err := u.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnmarshal, err)
	}
	return nil
}
