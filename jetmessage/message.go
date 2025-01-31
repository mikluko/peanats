package jetmessage

import (
	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
)

type TypedMessage[T any] interface {
	jetstream.Msg
	Payload() *T
}

var _ TypedMessage[any] = (*requestImpl[any])(nil)

type requestImpl[T any] struct {
	jetstream.Msg
	obj *T
}

func (r requestImpl[T]) Payload() *T { return r.obj }

// NewMessage is TypedMessage factory function
func NewMessage[T any](codec peanats.Codec, msg jetstream.Msg) (TypedMessage[T], error) {
	obj := new(T)
	err := codec.Decode(msg.Data(), obj)
	if err != nil {
		return nil, err
	}
	return requestImpl[T]{msg, obj}, nil
}
