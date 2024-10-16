package jetbucket

import (
	"errors"
	"io"

	"github.com/nats-io/nats.go/jetstream"
)

type watcher interface {
	jetstream.KeyWatcher
}

var ErrInitialValuesOver = errors.New("initial values done")

type Watcher[T any] interface {
	Next() (jetstream.KeyValueEntry, *T, error)
	Stop() error
}

func NewWatcher[T any](w watcher) Watcher[T] {
	return &watcherImpl[T]{
		w: w,
	}
}

type watcherImpl[T any] struct {
	w watcher
}

func (w *watcherImpl[T]) Next() (jetstream.KeyValueEntry, *T, error) {
	for ent := range w.w.Updates() {
		// nats will send a nil entry when it has received all initial values
		if ent == nil {
			return nil, nil, ErrInitialValuesOver
		}
		if ent.Operation() != jetstream.KeyValuePut {
			return ent, nil, nil
		}
		val := new(T)
		err := unmarshal(any(val), ent.Value())
		if err != nil {
			return nil, nil, err
		}
		return ent, val, nil
	}
	return nil, nil, io.EOF
}

func (w *watcherImpl[T]) Stop() error {
	return w.w.Stop()
}
