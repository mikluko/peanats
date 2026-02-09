package bucket

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
)

var (
	ErrInitialValuesOver = errors.New("initial values done")
	ErrDone              = errors.New("updates done")
)

type Watcher[T any] interface {
	Next() (Entry[T], error)
	Stop() error
}

func NewWatcher[T any](w jetstream.KeyWatcher, opts ...WatcherOption) Watcher[T] {
	p := bucketWatcherParams{}
	for _, opt := range opts {
		opt(&p)
	}
	return &watcherImpl[T]{
		watcher:  w,
		prefix:   p.prefix,
		metaOnly: p.metaOnly,
	}
}

type bucketWatcherParams struct {
	prefix       string
	metaOnly     bool
	upstreamOpts []jetstream.WatchOpt
}

type WatcherOption func(params *bucketWatcherParams)

func WatcherPrefix(prefix string) WatcherOption {
	return func(params *bucketWatcherParams) {
		params.prefix = prefix
	}
}

func WatcherMetaOnly() WatcherOption {
	return func(params *bucketWatcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.MetaOnly())
		params.metaOnly = true
	}
}

func WatcherUpdatesOnly() WatcherOption {
	return func(params *bucketWatcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.UpdatesOnly())
	}
}

func WatcherIgnoreDeletes() WatcherOption {
	return func(params *bucketWatcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.IgnoreDeletes())
	}
}

func WatcherIncludeHistory() WatcherOption {
	return func(params *bucketWatcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.IncludeHistory())
	}
}

func WatcherResumeFromRevision(revision uint64) WatcherOption {
	return func(params *bucketWatcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.ResumeFromRevision(revision))
	}
}

type watcherImpl[T any] struct {
	watcher     jetstream.KeyWatcher
	watcherOpts []jetstream.WatchOpt
	prefix      string
	metaOnly    bool
}

func (w *watcherImpl[T]) deprefixed(key string) string {
	if w.prefix == "" {
		return key
	}
	return strings.TrimPrefix(key, fmt.Sprintf("%s.", w.prefix))
}

func (w *watcherImpl[T]) Next() (_ Entry[T], err error) {
	for raw := range w.watcher.Updates() {
		// nats will send a nil entry when it has received all initial values
		if raw == nil {
			return nil, ErrInitialValuesOver
		}
		v := entryImpl[T]{
			entry: raw,
			key:   w.deprefixed(raw.Key()),
		}
		if raw.Operation() != jetstream.KeyValuePut {
			return v, nil
		}
		if w.metaOnly {
			return v, nil
		}
		v.header, v.value, err = decodeBucketEntryHeader[T](raw.Value())
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, ErrDone
}

func (w *watcherImpl[T]) Stop() error {
	return w.watcher.Stop()
}

type EntryHandler[T any] interface {
	HandleEntry(context.Context, Entry[T]) error
}

type EntryHandlerFunc[T any] func(context.Context, Entry[T]) error

func (f EntryHandlerFunc[T]) HandleEntry(ctx context.Context, e Entry[T]) error {
	return f(ctx, e)
}

type WatchOption func(params *watchParams)

type watchParams struct {
	disp peanats.Dispatcher
}

// WatchDispatcher sets the dispatcher.
func WatchDispatcher(disp peanats.Dispatcher) WatchOption {
	return func(p *watchParams) {
		p.disp = disp
	}
}

func Watch[T any](ctx context.Context, w Watcher[T], h EntryHandler[T], opts ...WatchOption) error {
	params := watchParams{
		disp: peanats.DefaultDispatcher,
	}
	for _, opt := range opts {
		opt(&params)
	}
	for {
		e, err := w.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if errors.Is(err, ErrInitialValuesOver) {
				continue
			}
			return err
		}
		params.disp.Dispatch(func() error {
			return h.HandleEntry(ctx, e)
		})
	}
}
