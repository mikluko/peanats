package publisher

import (
	"context"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
)

// RawPublisher is the dependency interface for low-level message publishing.
// transport.Conn satisfies this via structural typing.
type RawPublisher interface {
	Publish(context.Context, peanats.Msg) error
}

type PublishOption func(*PublishParams)

type PublishParams struct {
	Header          peanats.Header
	ContentType     codec.ContentType
	ContentEncoding codec.ContentEncoding
}

// WithHeader sets the header for the message.
// It can be used multiple times, but each time it will overwrite the previous value completely.
func WithHeader(header peanats.Header) PublishOption {
	return func(p *PublishParams) {
		p.Header = header
	}
}

// WithContentType sets the content type for the message.
func WithContentType(c codec.ContentType) PublishOption {
	return func(p *PublishParams) {
		p.ContentType = c
	}
}

// WithContentEncoding sets the compression algorithm for the message.
func WithContentEncoding(e codec.ContentEncoding) PublishOption {
	return func(p *PublishParams) {
		p.ContentEncoding = e
	}
}

type Publisher interface {
	Publish(context.Context, string, any, ...PublishOption) error
}

func New(pub RawPublisher) Publisher {
	return &publisherImpl{pub: pub}
}

type publisherImpl struct {
	pub RawPublisher
}

func (i *publisherImpl) Publish(ctx context.Context, subj string, v any, opts ...PublishOption) error {
	p := PublishParams{
		Header:      make(peanats.Header),
		ContentType: codec.JSON,
	}
	for _, o := range opts {
		o(&p)
	}
	p.Header.Set(codec.HeaderContentType, p.ContentType.String())
	if p.ContentEncoding != 0 {
		codec.SetContentEncoding(p.Header, p.ContentEncoding)
	}
	data, err := codec.MarshalHeader(v, p.Header)
	if err != nil {
		return err
	}
	return i.pub.Publish(ctx, msg{subj, p.Header, data})
}

type msg struct {
	subject string
	header  peanats.Header
	data    []byte
}

func (m msg) Subject() string {
	return m.subject
}

func (m msg) Header() peanats.Header {
	return m.header
}

func (m msg) Data() []byte {
	return m.data
}
