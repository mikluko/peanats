package peanats

import (
	"context"
	"net/textproto"

	natslib "github.com/nats-io/nats.go"
)

type PublishOption func(*PublishParams)

type PublishParams struct {
	Header      textproto.MIMEHeader
	ContentType ContentType
}

// PublishHeader sets the header for the message.
// It can be used multiple times, but each time it will overwrite the previous value completely.
func PublishHeader(header textproto.MIMEHeader) PublishOption {
	return func(p *PublishParams) {
		p.Header = header
	}
}

// PublishContentType sets the content type for the message.
func PublishContentType(c ContentType) PublishOption {
	return func(p *PublishParams) {
		p.ContentType = c
	}
}

type Publisher interface {
	Publish(context.Context, string, any, ...PublishOption) error
}

func NewPublisher(p publisher) Publisher {
	return &publisherImpl{p}
}

type publisher interface {
	PublishMsg(*natslib.Msg) error
}

type publisherImpl struct {
	pub publisher
}

func (i *publisherImpl) Publish(_ context.Context, subj string, v any, opts ...PublishOption) error {
	p := PublishParams{
		Header:      make(textproto.MIMEHeader),
		ContentType: ContentTypeJson,
	}
	for _, o := range opts {
		o(&p)
	}
	codec, err := CodecContentType(p.ContentType)
	if err != nil {
		return err
	}
	codec.SetHeader(p.Header)
	data, err := codec.Encode(v)
	if err != nil {
		return err
	}
	msg := &natslib.Msg{
		Subject: subj,
		Header:  natslib.Header(p.Header),
		Data:    data,
	}
	return i.pub.PublishMsg(msg)
}
