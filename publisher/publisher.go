package publisher

import (
	"context"

	"github.com/mikluko/peanats"
)

type PublishOption func(*PublishParams)

type PublishParams struct {
	Header      peanats.Header
	ContentType peanats.ContentType
}

// WithHeader sets the header for the message.
// It can be used multiple times, but each time it will overwrite the previous value completely.
func WithHeader(header peanats.Header) PublishOption {
	return func(p *PublishParams) {
		p.Header = header
	}
}

// WithContentType sets the content type for the message.
func WithContentType(c peanats.ContentType) PublishOption {
	return func(p *PublishParams) {
		p.ContentType = c
	}
}

type Publisher interface {
	Publish(context.Context, string, any, ...PublishOption) error
}

func New(pub peanats.Publisher) Publisher {
	return &publisherImpl{pub: pub}
}

type publisherImpl struct {
	pub peanats.Publisher
}

func (i *publisherImpl) Publish(ctx context.Context, subj string, v any, opts ...PublishOption) error {
	p := PublishParams{
		Header:      make(peanats.Header),
		ContentType: peanats.ContentTypeJson,
	}
	for _, o := range opts {
		o(&p)
	}
	codec, err := peanats.CodecContentType(p.ContentType)
	if err != nil {
		return err
	}
	codec.SetContentType(p.Header)
	data, err := codec.Marshal(v)
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
