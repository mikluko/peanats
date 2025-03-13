package peanats

import (
	"context"
	"net/textproto"

	natslib "github.com/nats-io/nats.go"
)

type RequestOption func(*requestParams)

type requestParams struct {
	header textproto.MIMEHeader
	ctype  ContentType
}

// RequestHeader sets the header for the message. It can be used multiple times, but each time it will
// overwrite the previous value completely.
func RequestHeader(header textproto.MIMEHeader) RequestOption {
	return func(p *requestParams) {
		p.header = header
	}
}

func RequestContentType(c ContentType) RequestOption {
	return func(p *requestParams) {
		p.ctype = c
	}
}

type Client[RQ, RS any] interface {
	Request(context.Context, string, *RQ, ...RequestOption) (Result[RS], error)
}

type natsRequester interface {
	RequestMsgWithContext(context.Context, *natslib.Msg) (*natslib.Msg, error)
}

func NewClient[RQ, RS any](nc natsRequester) Client[RQ, RS] {
	return &clientImpl[RQ, RS]{nc}
}

type clientImpl[RQ, RS any] struct {
	nc natsRequester
}

func (c *clientImpl[RQ, RS]) Request(ctx context.Context, subj string, rq *RQ, opts ...RequestOption) (Result[RS], error) {
	p := requestParams{
		header: make(textproto.MIMEHeader),
		ctype:  ContentTypeJson,
	}
	for _, opt := range opts {
		opt(&p)
	}
	codec, err := CodecContentType(p.ctype)
	if err != nil {
		return nil, err
	}
	codec.SetHeader(p.header)
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
	codec, err = CodecHeader(textproto.MIMEHeader(msg.Header))
	if err != nil {
		return nil, err
	}
	err = codec.Decode(msg.Data, rs)
	if err != nil {
		return nil, err
	}
	return &resultImpl[RS]{header: textproto.MIMEHeader(msg.Header), payload: rs}, nil
}

type Result[T any] interface {
	Header() textproto.MIMEHeader
	Payload() *T
}

type resultImpl[T any] struct {
	header  textproto.MIMEHeader
	payload *T
}

func (r *resultImpl[T]) Header() textproto.MIMEHeader {
	return r.header
}

func (r *resultImpl[T]) Payload() *T {
	return r.payload
}
