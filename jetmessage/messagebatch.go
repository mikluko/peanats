package jetmessage

import (
	"context"
	"io"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
)

type TypedMessageBatch[T any] interface {
	jetstream.MessageBatch
	Next(ctx context.Context) (TypedMessage[T], error)
}

func NewMessageBatch[T any](codec peanats.Codec, batch jetstream.MessageBatch) TypedMessageBatch[T] {
	return &messageBatchImpl[T]{batch, codec}
}

type messageBatchImpl[T any] struct {
	jetstream.MessageBatch
	codec peanats.Codec
}

func (r *messageBatchImpl[T]) Next(ctx context.Context) (TypedMessage[T], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-r.Messages():
		if !ok {
			return nil, io.EOF
		}
		return NewMessage[T](r.codec, msg)
	}
}
