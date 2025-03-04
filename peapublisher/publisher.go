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

func WithHeader(k, v string) Option {
	return func(p *params) {
		p.header.Add(k, v)
	}
}

func WithContentType(v string) Option {
	return WithHeader(peanats.HeaderContentType, v)
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
	codec := peanats.ChooseCodec(p.header)
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
