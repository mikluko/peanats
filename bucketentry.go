package peanats

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type BucketEntry[T any] interface {
	Bucket() string
	Key() string
	Header() textproto.MIMEHeader
	Value() *T
	Revision() uint64
	Created() time.Time
	Delta() uint64
	Operation() jetstream.KeyValueOp
}

type PutBucketEntry[T any] interface {
	Key() string
	Header() textproto.MIMEHeader
	Value() *T
}

type UpdateBucketEntry[T any] interface {
	Key() string
	Header() textproto.MIMEHeader
	Value() *T
	Revision() uint64
}

var _ BucketEntry[any] = (*entryImpl[any])(nil)

type entry interface {
	jetstream.KeyValueEntry
}

type entryImpl[T any] struct {
	entry
	value  *T
	header textproto.MIMEHeader
	key    string
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

const bucketEntryHeaderBoundary = "--"

func encodeBucketEntryHeader[T any](h textproto.MIMEHeader, v *T) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	if err := w.SetBoundary(bucketEntryHeaderBoundary); err != nil {
		panic(err)
	}
	c, err := CodecHeader(h)
	if err != nil {
		return nil, err
	}
	c.SetHeader(h)
	p, err := c.Encode(v)
	if err != nil {
		return nil, err
	}
	q, _ := w.CreatePart(h)
	_, _ = q.Write(p)
	return buf.Bytes(), nil
}

func decodeBucketEntryHeader[T any](b []byte) (h textproto.MIMEHeader, v *T, err error) {
	r := multipart.NewReader(bytes.NewReader(b), bucketEntryHeaderBoundary)
	p, err := r.NextPart()
	if err != nil {
		return nil, nil, err
	}
	h = p.Header
	c, err := CodecHeader(h)
	if err != nil {
		return nil, nil, err
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(p)
	v = new(T)
	err = c.Decode(buf.Bytes(), v)
	if err != nil {
		return nil, nil, err
	}
	return h, v, nil
}
