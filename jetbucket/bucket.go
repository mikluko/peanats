package jetbucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

type bucket interface {
	jetstream.KeyValue
}

type Bucket[T any] interface {
	Get(ctx context.Context, key string) (Entry[T], error)
	Put(ctx context.Context, key string, mod *T) (rev uint64, err error)
	Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error
	Watch(ctx context.Context, match string, opts ...jetstream.WatchOpt) (Watcher[T], error)
	WatchAll(ctx context.Context, opts ...jetstream.WatchOpt) (Watcher[T], error)
}

func NewBucket[T any](bucket bucket, opts ...BucketOption) Bucket[T] {
	if _, ok := any((*T)(nil)).(marshaler); !ok {
		panic("model type doesn't implement marshaler")
	}
	if _, ok := any((*T)(nil)).(unmarshaler); !ok {
		panic("model type doesn't implement unmarshaler")
	}
	params := &bucketParams{}
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
}

type BucketOption func(opts *bucketParams)

func WithKeyPrefix(prefix string) BucketOption {
	return func(params *bucketParams) {
		params.prefix = prefix
	}
}

// Bucket is a typed key-value store
type bucketImpl[T any] struct {
	bucket bucket
	prefix string
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

func (s *bucketImpl[T]) Get(ctx context.Context, key string) (Entry[T], error) {
	raw, err := s.bucket.Get(ctx, s.prefixed(key))
	if err != nil {
		return nil, err
	}
	v := entryImpl[T]{
		entry: raw,
		key:   s.deprefixed(raw.Key()),
	}
	if raw.Operation() != jetstream.KeyValuePut {
		return v, nil
	}
	v.value = new(T)
	err = unmarshal(any(v.value), raw.Value())
	if err != nil {
		return v, err
	}
	return v, nil
}

func (s *bucketImpl[T]) Put(ctx context.Context, key string, mod *T) (uint64, error) {
	b, err := marshal(any(mod).(marshaler))
	if err != nil {
		return 0, err
	}
	return s.bucket.Put(ctx, s.prefixed(key), b)
}

func (s *bucketImpl[T]) Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error {
	return s.bucket.Delete(ctx, s.prefixed(key), opts...)
}

func (s *bucketImpl[T]) Watch(ctx context.Context, match string, opts ...jetstream.WatchOpt) (Watcher[T], error) {
	w, err := s.bucket.Watch(ctx, s.prefixed(match), opts...)
	if err != nil {
		return nil, err
	}
	return NewWatcher[T](w, WatcherPrefix(s.prefix)), nil
}

func (s *bucketImpl[T]) WatchAll(ctx context.Context, opts ...jetstream.WatchOpt) (Watcher[T], error) {
	return s.Watch(ctx, ">", opts...)
}
