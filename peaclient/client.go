package peaclient

import (
	"context"

	natslib "github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

type Option func(*params)

type params struct {
	header peanats.Header
}

func WithHeader(k, v string) Option {
	return func(p *params) {
		p.header.Add(k, v)
	}
}

func WithContentType(v string) Option {
	return WithHeader(peanats.HeaderContentType, v)
}

type Client[RQ, RS any] interface {
	Request(context.Context, string, *RQ, ...Option) (Result[RS], error)
}

type nats interface {
	RequestMsgWithContext(context.Context, *natslib.Msg) (*natslib.Msg, error)
}

func New[RQ, RS any](nc nats) Client[RQ, RS] {
	return &clientImpl[RQ, RS]{nc}
}

type clientImpl[RQ, RS any] struct {
	nc nats
}

func (c *clientImpl[RQ, RS]) Request(ctx context.Context, subj string, rq *RQ, opts ...Option) (Result[RS], error) {
	p := params{header: make(peanats.Header)}
	for _, opt := range opts {
		opt(&p)
	}
	codec := peanats.ChooseCodec(p.header)
	data, err := codec.Encode(rq)
	if err != nil {
		return nil, err
	}
	msg := &natslib.Msg{
		Subject: subj,
		Header:  natslib.Header(p.header),
		Data:    data,
	}
	msg, err = c.nc.RequestMsgWithContext(ctx, msg)
	if err != nil {
		return nil, err
	}
	rs := new(RS)
	codec = peanats.ChooseCodec(peanats.Header(msg.Header))
	err = codec.Decode(msg.Data, rs)
	if err != nil {
		return nil, err
	}
	return &resultImpl[RS]{header: peanats.Header(msg.Header), payload: rs}, nil
}

type Result[T any] interface {
	Header() peanats.Header
	Payload() *T
}

type resultImpl[T any] struct {
	header  peanats.Header
	payload *T
}

func (r *resultImpl[T]) Header() peanats.Header {
	return r.header
}

func (r *resultImpl[T]) Payload() *T {
	return r.payload
}
