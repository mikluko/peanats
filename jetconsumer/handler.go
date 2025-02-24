package jetconsumer

import (
	"context"
	"net/http"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/jetmessage"
)

type Handler interface {
	Serve(ctx context.Context, msg jetstream.Msg) error
}

type HandlerFunc func(ctx context.Context, msg jetstream.Msg) error

func (f HandlerFunc) Serve(ctx context.Context, msg jetstream.Msg) error { return f(ctx, msg) }

type TypedHandler[T any] interface {
	Serve(context.Context, jetmessage.TypedMessage[T]) error
}

type TypedHandlerFunc[T any] func(context.Context, jetmessage.TypedMessage[T]) error

func (f TypedHandlerFunc[T]) Serve(ctx context.Context, msg jetmessage.TypedMessage[T]) error {
	return f(ctx, msg)
}

func HandleType[T any](c peanats.Codec, h TypedHandler[T]) Handler {
	return messageHandlerImpl[T]{c, h}
}

type messageHandlerImpl[T any] struct {
	c peanats.Codec
	h TypedHandler[T]
}

func (h messageHandlerImpl[T]) Serve(ctx context.Context, msg jetstream.Msg) (err error) {
	tm, err := jetmessage.NewMessage[T](h.c, msg)
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
