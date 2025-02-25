package jetbucket

import (
	"net/textproto"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type entry interface {
	jetstream.KeyValueEntry
}

type Entry[T any] interface {
	Bucket() string
	Key() string
	Header() textproto.MIMEHeader
	Value() *T
	Revision() uint64
	Created() time.Time
	Delta() uint64
	Operation() jetstream.KeyValueOp
}

var _ Entry[any] = (*entryImpl[any])(nil)

type entryImpl[T any] struct {
	entry
	key    string
	value  *T
	header textproto.MIMEHeader
}

func (e entryImpl[T]) Key() string {
	return e.key
}

func (e entryImpl[T]) Header() textproto.MIMEHeader {
	return e.header
}

func (e entryImpl[T]) RawValue() []byte {
	return e.entry.Value()
}

func (e entryImpl[T]) Value() *T {
	return e.value
}
