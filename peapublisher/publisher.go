package peapublisher

import (
	"context"

	"github.com/mikluko/peanats/xenc"
	"github.com/mikluko/peanats/xmsg"
	"github.com/mikluko/peanats/xnats"
)

type PublishOption func(*PublishParams)

type PublishParams struct {
	Header      xmsg.Header
	ContentType xenc.ContentType
}

// WithHeader sets the header for the message.
// It can be used multiple times, but each time it will overwrite the previous value completely.
func WithHeader(header xmsg.Header) PublishOption {
	return func(p *PublishParams) {
		p.Header = header
	}
}

// WithContentType sets the content type for the message.
func WithContentType(c xenc.ContentType) PublishOption {
	return func(p *PublishParams) {
		p.ContentType = c
	}
}

type Publisher interface {
	Publish(context.Context, string, any, ...PublishOption) error
}

func New(pub xnats.Publisher) Publisher {
	return &publisherImpl{pub: pub}
}

type publisherImpl struct {
	pub xnats.Publisher
}

func (i *publisherImpl) Publish(ctx context.Context, subj string, v any, opts ...PublishOption) error {
	p := PublishParams{
		Header:      make(xmsg.Header),
		ContentType: xenc.ContentTypeJson,
	}
	for _, o := range opts {
		o(&p)
	}
	codec, err := xenc.CodecContentType(p.ContentType)
	if err != nil {
		return err
	}
	codec.SetHeader(p.Header)
	data, err := codec.Encode(v)
	if err != nil {
		return err
	}
	return i.pub.Publish(ctx, msg{subj, p.Header, data})
}

type msg struct {
	subject string
	header  xmsg.Header
	data    []byte
}

func (m msg) Subject() string {
	return m.subject
}

func (m msg) Header() xmsg.Header {
	return m.header
}

func (m msg) Data() []byte {
	return m.data
}
