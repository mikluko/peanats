package peanats

import (
	"context"
	"errors"
	"fmt"

	"github.com/mikluko/peanats/internal/xargpool"
)

type Argument[T any] interface {
	Message
	Value() *T
}

// ArgumentHandler interface defines a typed handler.
type ArgumentHandler[T any] interface {
	HandleArgument(context.Context, Argument[T]) error
}

// ArgumentHandlerFunc is an adapter to allow the use of ordinary functions as Handler.
type ArgumentHandlerFunc[T any] func(context.Context, Argument[T]) error

func (f ArgumentHandlerFunc[T]) HandleArgument(ctx context.Context, a Argument[T]) error {
	return f(ctx, a)
}

var ErrArgumentDecodeFailed = errors.New("failed to decode message into argument")

func ArgumentMessageHandler[T any](h ArgumentHandler[T]) MessageHandler {
	pool := xargpool.New[T]()
	return MessageHandlerFunc(func(ctx context.Context, m Message) error {
		codec, err := CodecContentType(ContentTypeHeader(m.Header()))
		if err != nil {
			return err
		}

		x := pool.Acquire(ctx)
		v := x.Value()
		defer x.Release()

		err = codec.Decode(m.Data(), v)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrArgumentDecodeFailed, err)
		}

		if mjs, ok := m.(MessageJetstream); ok {
			return h.HandleArgument(ctx, argumentJetstreamImpl[T]{mjs, v})
		} else {
			return h.HandleArgument(ctx, argumentImpl[T]{m, v})
		}
	})
}

var _ Argument[any] = (*argumentImpl[any])(nil)

type argumentImpl[T any] struct {
	Message
	v *T
}

func (a argumentImpl[T]) Value() *T {
	return a.v
}

type argumentJetstreamImpl[T any] struct {
	MessageJetstream
	v *T
}

func (a argumentJetstreamImpl[T]) Value() *T {
	return a.v
}
