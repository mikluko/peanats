package peanats

import (
	"context"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

type Bucket[T any] interface {
	Get(ctx context.Context, key string) (BucketEntry[T], error)
	GetRevision(ctx context.Context, key string, rev uint64) (BucketEntry[T], error)
	Put(ctx context.Context, entry PutBucketEntry[T]) (uint64, error)
	Update(ctx context.Context, entry UpdateBucketEntry[T]) (uint64, error)
	Delete(ctx context.Context, key string, opts ...DeleteOption) error
	Watch(ctx context.Context, match string, opts ...BucketWatcherOption) (BucketWatcher[T], error)
	WatchAll(ctx context.Context, opts ...BucketWatcherOption) (BucketWatcher[T], error)
}

func NewBucket[T any](bucket jetstream.KeyValue, opts ...BucketOption) Bucket[T] {
	params := &bucketParams{
		codec: JsonCodec{},
	}
	for _, opt := range opts {
		opt(params)
	}
	return &bucketImpl[T]{
		bucket: bucket,
		prefix: params.prefix,
	}
}

type bucketParams struct {
	prefix string
	codec  Codec
}

type BucketOption func(opts *bucketParams)

func BucketKeyPrefix(prefix string) BucketOption {
	return func(params *bucketParams) {
		params.prefix = prefix
	}
}

// Bucket is a typed key-value store
type bucketImpl[T any] struct {
	bucket jetstream.KeyValue
	prefix string
	codec  Codec
}

func (s *bucketImpl[T]) prefixed(key string) string {
	if s.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s.%s", s.prefix, key)
}

func (s *bucketImpl[T]) deprefixed(key string) string {
	if s.prefix == "" {
		return key
	}
	return strings.TrimPrefix(key, fmt.Sprintf("%s.", s.prefix))
}

func (s *bucketImpl[T]) get(raw jetstream.KeyValueEntry) (_ BucketEntry[T], err error) {
	v := entryImpl[T]{
		entry: raw,
		key:   s.deprefixed(raw.Key()),
	}
	if raw.Operation() != jetstream.KeyValuePut {
		return v, nil
	}
	v.header, v.value, err = decodeBucketEntryHeader[T](raw.Value())
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *bucketImpl[T]) Get(ctx context.Context, key string) (BucketEntry[T], error) {
	raw, err := s.bucket.Get(ctx, s.prefixed(key))
	if err != nil {
		return nil, err
	}
	return s.get(raw)
}

func (s *bucketImpl[T]) GetRevision(ctx context.Context, key string, rev uint64) (BucketEntry[T], error) {
	raw, err := s.bucket.GetRevision(ctx, s.prefixed(key), rev)
	if err != nil {
		return nil, err
	}
	return s.get(raw)
}

func (s *bucketImpl[T]) Put(ctx context.Context, entry PutBucketEntry[T]) (uint64, error) {
	b, err := encodeBucketEntryHeader(entry.Header(), entry.Value())
	if err != nil {
		return 0, err
	}
	return s.bucket.Put(ctx, s.prefixed(entry.Key()), b)
}

func (s *bucketImpl[T]) Update(ctx context.Context, entry UpdateBucketEntry[T]) (uint64, error) {
	b, err := encodeBucketEntryHeader(entry.Header(), entry.Value())
	if err != nil {
		return 0, err
	}
	return s.bucket.Update(ctx, s.prefixed(entry.Key()), b, entry.Revision())
}

type bucketImplUpdateParams struct {
	header textproto.MIMEHeader
}

type UpdateOption func(params *bucketImplUpdateParams)

func UpdateWithHeader(header textproto.MIMEHeader) UpdateOption {
	return func(params *bucketImplUpdateParams) {
		params.header = header
	}
}

type DeleteOption = jetstream.KVDeleteOpt

func (s *bucketImpl[T]) Delete(ctx context.Context, key string, opts ...DeleteOption) error {
	return s.bucket.Delete(ctx, s.prefixed(key), opts...)
}

func (s *bucketImpl[T]) Watch(ctx context.Context, match string, opts ...BucketWatcherOption) (BucketWatcher[T], error) {
	params := bucketWatcherParams{}
	for _, opt := range opts {
		opt(&params)
	}
	w, err := s.bucket.Watch(ctx, s.prefixed(match), params.upstreamOpts...)
	if err != nil {
		return nil, err
	}
	return NewBucketWatcher[T](w, append(opts, BucketWatcherPrefix(s.prefix))...), nil
}

func (s *bucketImpl[T]) WatchAll(ctx context.Context, opts ...BucketWatcherOption) (BucketWatcher[T], error) {
	return s.Watch(ctx, ">", opts...)
}
