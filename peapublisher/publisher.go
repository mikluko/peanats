package peapublisher

import (
	"context"

	natslib "github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

type Option func(*params)

type params struct {
	header peanats.Header
	ctype  peanats.ContentType
}

// WithHeader sets the header for the message. It can be used multiple times, but each time it will
// overwrite the previous value completely.
func WithHeader(header peanats.Header) Option {
	return func(p *params) {
		p.header = header
	}
}

func WithContentType(c peanats.ContentType) Option {
	return func(p *params) {
		p.ctype = c
	}
}

type Publisher interface {
	Publish(context.Context, string, any, ...Option) error
}

func New(p nats) Publisher {
	return &publisherImpl{p}
}

type nats interface {
	PublishMsg(*natslib.Msg) error
}

type publisherImpl struct {
	pub nats
}

func (i *publisherImpl) Publish(_ context.Context, subj string, v any, opts ...Option) error {
	p := params{
		header: peanats.Header{},
		ctype:  peanats.ContentTypeJson,
	}
	for _, o := range opts {
		o(&p)
	}
	codec, err := peanats.CodecContentType(p.ctype)
	if err != nil {
		return err
	}
	data, err := codec.Encode(v)
	if err != nil {
		return err
	}
	codec.SetHeader(p.header)
	msg := &natslib.Msg{
		Subject: subj,
		Header:  natslib.Header(p.header),
		Data:    data,
	}
	return i.pub.PublishMsg(msg)
}
