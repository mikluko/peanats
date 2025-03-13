package peaclient

import (
	"context"
	"errors"

	"github.com/mikluko/peanats/xarg"
	"github.com/mikluko/peanats/xenc"
	"github.com/mikluko/peanats/xmsg"
	"github.com/mikluko/peanats/xnats"
)

type RequestOption func(*requestParams)

type requestParams struct {
	header xmsg.Header
	ctype  xenc.ContentType
}

// RequestHeader sets the header for the message. It can be used multiple times, but each time it will
// overwrite the previous value completely.
func RequestHeader(header xmsg.Header) RequestOption {
	return func(p *requestParams) {
		p.header = header
	}
}

func RequestContentType(c xenc.ContentType) RequestOption {
	return func(p *requestParams) {
		p.ctype = c
	}
}

type Client[RQ, RS any] interface {
	Request(context.Context, string, *RQ, ...RequestOption) (Response[RS], error)
}

func New[RQ, RS any](nc xnats.Connection) Client[RQ, RS] {
	return &clientImpl[RQ, RS]{nc}
}

type clientImpl[RQ, RS any] struct {
	nc xnats.Connection
}

func (c *clientImpl[RQ, RS]) Request(ctx context.Context, subj string, rq *RQ, opts ...RequestOption) (Response[RS], error) {
	p := requestParams{
		header: make(xmsg.Header),
		ctype:  xenc.ContentTypeJson,
	}
	for _, opt := range opts {
		opt(&p)
	}
	codec, err := xenc.CodecContentType(p.ctype)
	if err != nil {
		return nil, err
	}
	codec.SetHeader(p.header)
	data, err := codec.Encode(rq)
	if err != nil {
		return nil, err
	}
	msg, err := c.nc.Request(ctx, requestMessageImpl{subj: subj, header: p.header, data: data})
	if err != nil {
		return nil, err
	}
	rs := new(RS)
	codec, err = xenc.CodecHeader(msg.Header())
	if err != nil {
		return nil, err
	}
	err = codec.Decode(msg.Data(), rs)
	if err != nil {
		return nil, err
	}
	return &responseImpl[RS]{header: msg.Header(), payload: rs}, nil
}

func (c *clientImpl[RQ, RS]) Receiver(ctx context.Context, subj string, rq *RQ, opts ...RequestOption) (ResponseReceiver[RS], error) {
	panic("not implemented")
}

type requestMessageImpl struct {
	subj   string
	header xmsg.Header
	data   []byte
}

func (r requestMessageImpl) Subject() string {
	return r.subj
}

func (r requestMessageImpl) Header() xmsg.Header {
	return r.header
}

func (r requestMessageImpl) Data() []byte {
	return r.data
}

type Response[T any] interface {
	Header() xmsg.Header
	Value() *T
}

type responseImpl[T any] struct {
	header  xmsg.Header
	payload *T
}

func (r *responseImpl[T]) Header() xmsg.Header {
	return r.header
}

func (r *responseImpl[T]) Value() *T {
	return r.payload
}

type ResponseReceiver[T any] interface {
	Next(context.Context) (Response[T], error)
	Handle(context.Context, xarg.ArgHandler[T]) error
}

type responseReceiverImpl[T any] struct{}

func (r responseReceiverImpl[T]) Next(ctx context.Context) (Response[T], error) {
	return nil, errors.New("not implemented")
}

func (r responseReceiverImpl[T]) Handle(ctx context.Context, h xarg.ArgHandler[T]) error {
	return errors.New("not implemented")
}
