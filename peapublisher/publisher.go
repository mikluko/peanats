package peapublisher

import (
	"context"

	natslib "github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

type Option func(*params)

type params struct {
	header peanats.Header
}

// WithHeader sets the header for the message. It can be used multiple times, but each time it will
// overwrite the previous value completely.
func WithHeader(header peanats.Header) Option {
	return func(p *params) {
		p.header = header
	}
}

func WithContentType(v string) Option {
	return func(p *params) {
		if p.header == nil {
			p.header = peanats.Header{}
		}
		p.header.Set(peanats.HeaderContentType, v)
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
	}
	for _, o := range opts {
		o(&p)
	}
	codec, err := peanats.CodecHeader(p.header)
	if err != nil {
		return err
	}
	data, err := codec.Encode(v)
	if err != nil {
		return err
	}
	msg := &natslib.Msg{
		Subject: subj,
		Header:  natslib.Header(p.header),
		Data:    data,
	}
	return i.pub.PublishMsg(msg)
}
