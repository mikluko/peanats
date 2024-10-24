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
		watcher: w,
		prefix:  p.prefix,
	}
}

type watcherParams struct {
	prefix string
}

type WatcherOption func(params *watcherParams)

func WatcherPrefix(prefix string) WatcherOption {
	return func(params *watcherParams) {
		params.prefix = prefix
	}
}

type watcherImpl[T any] struct {
	watcher watcher
	prefix  string
}

func (w *watcherImpl[T]) Next() (Entry[T], error) {
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
		v.value = new(T)
		err := unmarshal(any(v.value), raw.Value())
		if err != nil {
			return v, err
		}
		return v, nil
	}
	return nil, io.EOF
}

func (w *watcherImpl[T]) Stop() error {
	return w.watcher.Stop()
}
