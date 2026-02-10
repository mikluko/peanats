package bucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
)

type Bucket[T any] interface {
	Get(ctx context.Context, key string) (Entry[T], error)
	GetRevision(ctx context.Context, key string, rev uint64) (Entry[T], error)
	GetLatestRevision(ctx context.Context, key string) (Entry[T], error)
	Put(ctx context.Context, entry PutEntry[T]) (uint64, error)
	Update(ctx context.Context, entry UpdateEntry[T]) (uint64, error)
	Delete(ctx context.Context, key string, opts ...DeleteOption) error
	History(ctx context.Context, key string, opts ...HistoryOption) ([]Entry[T], error)
	Watch(ctx context.Context, match string, opts ...WatcherOption) (Watcher[T], error)
	WatchAll(ctx context.Context, opts ...WatcherOption) (Watcher[T], error)
}

func NewBucket[T any](bucket jetstream.KeyValue, opts ...BucketOption) Bucket[T] {
	params := &bucketParams{}
	for _, opt := range opts {
		opt(params)
	}
	return &bucketImpl[T]{
		bucket:          bucket,
		prefix:          params.prefix,
		contentEncoding: params.contentEncoding,
	}
}

type bucketParams struct {
	prefix          string
	contentEncoding codec.ContentEncoding
}

type BucketOption func(opts *bucketParams)

func BucketKeyPrefix(prefix string) BucketOption {
	return func(params *bucketParams) {
		params.prefix = prefix
	}
}

// BucketContentEncoding sets the compression algorithm for bucket entries.
func BucketContentEncoding(e codec.ContentEncoding) BucketOption {
	return func(params *bucketParams) {
		params.contentEncoding = e
	}
}

// Bucket is a typed key-value store
type bucketImpl[T any] struct {
	bucket          jetstream.KeyValue
	prefix          string
	contentEncoding codec.ContentEncoding
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

func (s *bucketImpl[T]) get(raw jetstream.KeyValueEntry) (_ Entry[T], err error) {
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

func (s *bucketImpl[T]) Get(ctx context.Context, key string) (Entry[T], error) {
	raw, err := s.bucket.Get(ctx, s.prefixed(key))
	if err != nil {
		return nil, err
	}
	return s.get(raw)
}

func (s *bucketImpl[T]) GetRevision(ctx context.Context, key string, rev uint64) (Entry[T], error) {
	raw, err := s.bucket.GetRevision(ctx, s.prefixed(key), rev)
	if err != nil {
		return nil, err
	}
	return s.get(raw)
}

func (s *bucketImpl[T]) GetLatestRevision(ctx context.Context, key string) (Entry[T], error) {
	watcher, err := s.bucket.Watch(ctx, s.prefixed(key))
	if err != nil {
		return nil, err
	}
	defer watcher.Stop()

	entry := <-watcher.Updates()
	if entry == nil {
		return nil, jetstream.ErrKeyNotFound
	}
	return s.get(entry)
}

func (s *bucketImpl[T]) Put(ctx context.Context, entry PutEntry[T]) (uint64, error) {
	h := entry.Header()
	if s.contentEncoding != 0 {
		codec.SetContentEncoding(h, s.contentEncoding)
	}
	b, err := encodeBucketEntryHeader(h, entry.Value())
	if err != nil {
		return 0, err
	}
	return s.bucket.Put(ctx, s.prefixed(entry.Key()), b)
}

func (s *bucketImpl[T]) Update(ctx context.Context, entry UpdateEntry[T]) (uint64, error) {
	h := entry.Header()
	if s.contentEncoding != 0 {
		codec.SetContentEncoding(h, s.contentEncoding)
	}
	b, err := encodeBucketEntryHeader(h, entry.Value())
	if err != nil {
		return 0, err
	}
	return s.bucket.Update(ctx, s.prefixed(entry.Key()), b, entry.Revision())
}

type bucketImplUpdateParams struct {
	header peanats.Header
}

type UpdateOption func(params *bucketImplUpdateParams)

func UpdateWithHeader(header peanats.Header) UpdateOption {
	return func(params *bucketImplUpdateParams) {
		params.header = header
	}
}

type DeleteOption = jetstream.KVDeleteOpt

func (s *bucketImpl[T]) Delete(ctx context.Context, key string, opts ...DeleteOption) error {
	return s.bucket.Delete(ctx, s.prefixed(key), opts...)
}

type HistoryOption = jetstream.WatchOpt

func (s *bucketImpl[T]) History(ctx context.Context, key string, opts ...HistoryOption) ([]Entry[T], error) {
	rawEntries, err := s.bucket.History(ctx, s.prefixed(key), opts...)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry[T], len(rawEntries))
	for i, raw := range rawEntries {
		entries[i], err = s.get(raw)
		if err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func (s *bucketImpl[T]) Watch(ctx context.Context, match string, opts ...WatcherOption) (Watcher[T], error) {
	params := bucketWatcherParams{}
	for _, opt := range opts {
		opt(&params)
	}
	w, err := s.bucket.Watch(ctx, s.prefixed(match), params.upstreamOpts...)
	if err != nil {
		return nil, err
	}
	return NewWatcher[T](w, append(opts, WatcherPrefix(s.prefix))...), nil
}

func (s *bucketImpl[T]) WatchAll(ctx context.Context, opts ...WatcherOption) (Watcher[T], error) {
	return s.Watch(ctx, ">", opts...)
}
