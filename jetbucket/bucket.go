package jetbucket

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

type bucket interface {
	jetstream.KeyValue
}

type Bucket[T any] interface {
	Get(ctx context.Context, key string) (jetstream.KeyValueEntry, *T, error)
	Put(ctx context.Context, key string, mod *T) (rev uint64, err error)
	Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error
	Watch(ctx context.Context, match string, opts ...jetstream.WatchOpt) (Watcher[T], error)
	WatchAll(ctx context.Context, opts ...jetstream.WatchOpt) (Watcher[T], error)
}

func NewBucket[T any](bucket bucket) Bucket[T] {
	if _, ok := any((*T)(nil)).(marshaler); !ok {
		panic("model type doesn't implement marshaler")
	}
	if _, ok := any((*T)(nil)).(unmarshaler); !ok {
		panic("model type doesn't implement unmarshaler")
	}
	return &bucketImpl[T]{
		bucket: bucket,
	}
}

// Bucket is a typed key-value store
type bucketImpl[T any] struct {
	bucket bucket
}

func (s *bucketImpl[T]) Get(ctx context.Context, key string) (jetstream.KeyValueEntry, *T, error) {
	ent, err := s.bucket.Get(ctx, key)
	if err != nil {
		return nil, nil, err
	}
	if ent.Operation() != jetstream.KeyValuePut {
		return ent, nil, nil
	}
	val := new(T)
	err = unmarshal(any(val), ent.Value())
	if err != nil {
		return nil, nil, err
	}
	return ent, val, nil
}

func (s *bucketImpl[T]) Put(ctx context.Context, key string, mod *T) (uint64, error) {
	b, err := marshal(any(mod).(marshaler))
	if err != nil {
		return 0, err
	}
	return s.bucket.Put(ctx, key, b)
}

func (s *bucketImpl[T]) Delete(ctx context.Context, key string, opts ...jetstream.KVDeleteOpt) error {
	return s.bucket.Delete(ctx, key, opts...)
}

func (s *bucketImpl[T]) Watch(ctx context.Context, match string, opts ...jetstream.WatchOpt) (Watcher[T], error) {
	w, err := s.bucket.Watch(ctx, match, opts...)
	if err != nil {
		return nil, err
	}
	return NewWatcher[T](w), nil
}

func (s *bucketImpl[T]) WatchAll(ctx context.Context, opts ...jetstream.WatchOpt) (Watcher[T], error) {
	w, err := s.bucket.WatchAll(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return NewWatcher[T](w), nil
}
