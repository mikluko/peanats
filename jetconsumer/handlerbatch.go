package jetconsumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/jetmessage"
)

type BatchHandler interface {
	Serve(ctx context.Context, batch jetstream.MessageBatch) error
}

type BatchHandlerFunc func(ctx context.Context, batch jetstream.MessageBatch) error

func (f BatchHandlerFunc) Serve(ctx context.Context, batch jetstream.MessageBatch) error {
	return f(ctx, batch)
}

type TypedBatchHandler[T any] interface {
	Serve(context.Context, jetmessage.TypedMessageBatch[T]) error
}

type TypedBatchHandlerFunc[T any] func(context.Context, jetmessage.TypedMessageBatch[T]) error

func (f TypedBatchHandlerFunc[T]) Serve(ctx context.Context, batch jetmessage.TypedMessageBatch[T]) error {
	return f(ctx, batch)
}

func HandleBatchType[T any](c peanats.Codec, h TypedBatchHandler[T]) BatchHandler {
	return batchHandlerImpl[T]{c, h}
}

type batchHandlerImpl[T any] struct {
	c peanats.Codec
	h TypedBatchHandler[T]
}

func (h batchHandlerImpl[T]) Serve(ctx context.Context, raw jetstream.MessageBatch) (err error) {
	batch := jetmessage.NewMessageBatch[T](h.c, raw)
	return h.h.Serve(ctx, batch)
}
