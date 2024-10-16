package jetconsumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Handler[T any] interface {
	Serve(context.Context, *T) error
}

type MessageHandlerFunc func(ctx context.Context, msg jetstream.Msg) error

func (f MessageHandlerFunc) Serve(ctx context.Context, msg jetstream.Msg) error {
	return f(ctx, msg)
}

type MessageHandler interface {
	Serve(ctx context.Context, msg jetstream.Msg) error
}

func messageHandlerFrom[T any](h Handler[T]) MessageHandler {
	return messageHandlerImpl[T]{h}
}

type messageHandlerImpl[T any] struct {
	h Handler[T]
}

func (h messageHandlerImpl[T]) Serve(ctx context.Context, msg jetstream.Msg) (err error) {
	arg := new(T)
	err = protojson.Unmarshal(msg.Data(), any(arg).(proto.Message))
	if err != nil {
		return err
	}
	err = h.h.Serve(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}
