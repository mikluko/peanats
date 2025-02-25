package jetbucket

import (
	"errors"
	"io"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

type watcher interface {
	jetstream.KeyWatcher
}

var ErrInitialValuesOver = errors.New("initial values done")

type Watcher[T any] interface {
	Next() (Entry[T], error)
	Stop() error
}

func NewWatcher[T any](w watcher, opts ...WatcherOption) Watcher[T] {
	p := watcherParams{}
	for _, opt := range opts {
		opt(&p)
	}
	return &watcherImpl[T]{
		watcher:  w,
		prefix:   p.prefix,
		metaOnly: p.metaOnly,
	}
}

type watcherParams struct {
	prefix       string
	metaOnly     bool
	upstreamOpts []jetstream.WatchOpt
}

type WatcherOption func(params *watcherParams)

func WatcherPrefix(prefix string) WatcherOption {
	return func(params *watcherParams) {
		params.prefix = prefix
	}
}

func WatcherMetaOnly() WatcherOption {
	return func(params *watcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.MetaOnly())
		params.metaOnly = true
	}
}

func WatcherUpdatesOnly() WatcherOption {
	return func(params *watcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.UpdatesOnly())
	}
}

func WatcherIgnoreDeletes() WatcherOption {
	return func(params *watcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.IgnoreDeletes())
	}
}

func WatcherIncludeHistory() WatcherOption {
	return func(params *watcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.IncludeHistory())
	}
}

func WatcherResumeFromRevision(revision uint64) WatcherOption {
	return func(params *watcherParams) {
		params.upstreamOpts = append(params.upstreamOpts, jetstream.ResumeFromRevision(revision))
	}
}

type watcherImpl[T any] struct {
	watcher     watcher
	watcherOpts []jetstream.WatchOpt
	prefix      string
	metaOnly    bool
}

func (w *watcherImpl[T]) Next() (_ Entry[T], err error) {
	for raw := range w.watcher.Updates() {
		// nats will send a nil entry when it has received all initial values
		if raw == nil {
			return nil, ErrInitialValuesOver
		}
		v := entryImpl[T]{
			entry: raw,
			key:   raw.Key(),
		}
		if w.prefix != "" {
			v.key = strings.TrimPrefix(v.key, w.prefix+".")
		}
		if raw.Operation() != jetstream.KeyValuePut {
			return v, nil
		}
		if w.metaOnly {
			return v, nil
		}
		v.header, v.value, err = decode[T](raw.Value())
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return nil, io.EOF
}

func (w *watcherImpl[T]) Stop() error {
	return w.watcher.Stop()
}
