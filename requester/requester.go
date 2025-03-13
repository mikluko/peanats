package requester

import (
	"context"
	"io"

	"github.com/mikluko/peanats"
)

type RequestOption func(*requestParams)

type requestParams struct {
	header peanats.Header
	ctype  peanats.ContentType
	buffer uint
}

// RequestHeader sets the header for the message. It can be used multiple times, but each time it will
// overwrite the previous value completely.
func RequestHeader(header peanats.Header) RequestOption {
	return func(p *requestParams) {
		p.header = header
	}
}

func RequestContentType(c peanats.ContentType) RequestOption {
	return func(p *requestParams) {
		p.ctype = c
	}
}

type Requester[RQ, RS any] interface {
	Request(context.Context, string, *RQ, ...RequestOption) (Response[RS], error)
	ResponseReceiver(context.Context, string, *RQ, ...RequestOption) (ResponseReceiver[RS], error)
}

func New[RQ, RS any](nc peanats.Connection) Requester[RQ, RS] {
	return &clientImpl[RQ, RS]{nc}
}

type clientImpl[RQ, RS any] struct {
	nc peanats.Connection
}

func (c *clientImpl[RQ, RS]) Request(ctx context.Context, subj string, rq *RQ, opts ...RequestOption) (Response[RS], error) {
	p := requestParams{
		header: make(peanats.Header),
		ctype:  peanats.ContentTypeJson,
	}
	for _, opt := range opts {
		opt(&p)
	}
	codec, err := peanats.CodecContentType(p.ctype)
	if err != nil {
		return nil, err
	}
	codec.SetContentType(p.header)
	data, err := codec.Marshal(rq)
	if err != nil {
		return nil, err
	}
	msg, err := c.nc.Request(ctx, requestMessageImpl{subj: subj, header: p.header, data: data})
	if err != nil {
		return nil, err
	}
	rs := new(RS)
	codec, err = peanats.CodecHeader(msg.Header())
	if err != nil {
		return nil, err
	}
	err = codec.Unmarshal(msg.Data(), rs)
	if err != nil {
		return nil, err
	}
	return &responseImpl[RS]{header: msg.Header(), payload: rs}, nil
}

func (c *clientImpl[RQ, RS]) ResponseReceiver(ctx context.Context, subj string, rq *RQ, opts ...RequestOption) (ResponseReceiver[RS], error) {
	panic("not implemented")
}

type requestMessageImpl struct {
	subj   string
	header peanats.Header
	data   []byte
}

func (r requestMessageImpl) Subject() string {
	return r.subj
}

func (r requestMessageImpl) Header() peanats.Header {
	return r.header
}

func (r requestMessageImpl) Data() []byte {
	return r.data
}

type Response[T any] interface {
	Header() peanats.Header
	Value() *T
}

type responseImpl[T any] struct {
	header  peanats.Header
	payload *T
}

func (r *responseImpl[T]) Header() peanats.Header {
	return r.header
}

func (r *responseImpl[T]) Value() *T {
	return r.payload
}

type ResponseReceiver[T any] interface {
	Next(context.Context) (Response[T], error)
	Stop() error
}

type responseReceiverImpl[T any] struct {
	buf chan peanats.Msg
}

func (r responseReceiverImpl[T]) Next(ctx context.Context) (Response[T], error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-r.buf:
		if msg == nil {
			return nil, io.EOF
		}
		x := new(T)
		err := peanats.UnmarshalHeader(msg.Data(), x, msg.Header())
		if err != nil {
			return nil, err
		}
		return &responseImpl[T]{header: msg.Header(), payload: x}, nil
	}
}

func (r responseReceiverImpl[T]) Stop() error {
	close(r.buf)
	return nil
}
