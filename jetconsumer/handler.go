package jetconsumer

import (
	"context"
	"net/http"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
)

type Handler interface {
	Serve(ctx context.Context, msg jetstream.Msg) error
}

type HandlerFunc func(ctx context.Context, msg jetstream.Msg) error

func (f HandlerFunc) Serve(ctx context.Context, msg jetstream.Msg) error { return f(ctx, msg) }

type TypedHandler[T any] interface {
	Serve(context.Context, TypedMessage[T]) error
}

type TypedHandlerFunc[T any] func(peanats.TypedRequest[T]) error

func (f TypedHandlerFunc[T]) Serve(arg peanats.TypedRequest[T]) error { return f(arg) }

func Handle[T any](h TypedHandler[T]) Handler {
	return messageHandlerImpl[T]{h}
}

type messageHandlerImpl[T any] struct {
	h TypedHandler[T]
}

func (h messageHandlerImpl[T]) Serve(ctx context.Context, msg jetstream.Msg) (err error) {
	tm, err := message[T](msg)
	if err != nil {
		return err
	}
	return h.h.Serve(ctx, tm)
}

var NotFoundHandler = HandlerFunc(func(_ context.Context, _ jetstream.Msg) error {
	return peanats.Error{
		Code:    http.StatusNotFound,
		Message: "handler not found",
		Cause:   nil,
	}
})
