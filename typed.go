package peanats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TypedRequest[T any] interface {
	Context() context.Context
	Header() *nats.Header
	Payload() *T
}

type TypedPublisher[ResT any] interface {
	Header() *nats.Header
	Publish(*ResT) error
}

type TypedHandler[ArgT, ResT any] interface {
	Serve(TypedPublisher[ResT], TypedRequest[ArgT]) error
}

type TypedHandlerFunc[ArgT, ResT any] func(TypedPublisher[ResT], TypedRequest[ArgT]) error

func (f TypedHandlerFunc[ArgT, ResT]) Serve(pub TypedPublisher[ResT], req TypedRequest[ArgT]) error {
	return f(pub, req)
}

func Typed[ArgT, ResT any](c Codec, f TypedHandler[ArgT, ResT]) Handler {
	return HandlerFunc(func(pub Publisher, req Request) error {
		arg := new(ArgT)
		if err := c.Decode(req.Data(), arg); err != nil {
			return &Error{
				Code:    http.StatusBadRequest,
				Message: http.StatusText(http.StatusBadRequest),
				Cause:   err,
			}
		}
		return f.Serve(&typedPublisherImpl[ResT]{c, pub}, &typedRequestImpl[ArgT]{arg: arg})
	})
}

type typedRequestImpl[ArgT any] struct {
	req Request
	arg *ArgT
}

func (r *typedRequestImpl[ArgT]) Context() context.Context {
	return r.req.Context()
}

func (r *typedRequestImpl[ArgT]) Header() *nats.Header {
	return r.req.Header()
}

func (r *typedRequestImpl[ArgT]) Payload() *ArgT {
	return r.arg
}

type typedPublisherImpl[RS any] struct {
	codec Codec
	pub   Publisher
}

func (p *typedPublisherImpl[RS]) Header() *nats.Header {
	return p.pub.Header()
}

func (p *typedPublisherImpl[RS]) Publish(v *RS) error {
	data, err := p.codec.Encode(v)
	if err != nil {
		return &Error{
			Code:    http.StatusInternalServerError,
			Message: http.StatusText(http.StatusInternalServerError),
			Cause:   err,
		}
	}
	return p.pub.Publish(data)
}

type Codec interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, vPtr any) error
}

type JsonCodec struct{}

func (JsonCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JsonCodec) Decode(data []byte, vPtr any) error {
	return json.Unmarshal(data, vPtr)
}

type ProtojsonCodec struct{}

func (ProtojsonCodec) Encode(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return protojson.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (ProtojsonCodec) Decode(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return protojson.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}
