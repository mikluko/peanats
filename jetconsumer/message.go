package jetconsumer

import (
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

// message is TypedMessage factory function
func message[T any](msg jetstream.Msg) (TypedMessage[T], error) {
	obj := new(T)
	err := protojson.Unmarshal(msg.Data(), any(obj).(proto.Message))
	if err != nil {
		return nil, err
	}
	return requestImpl[T]{msg, obj}, nil
}
