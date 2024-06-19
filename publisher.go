package peanats

import (
	"errors"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	PublisherMsg
	Subject() string
	WithSubject(subject string) Publisher
	Header() *nats.Header
	Publish(data []byte) error
}

func NewPublisher(p PublisherMsg) Publisher {
	return NewPublisherWithSubject(p, "")
}

func NewPublisherWithSubject(p PublisherMsg, subject string) Publisher {
	return &publisherImpl{
		PublisherMsg: p,
		subject:      subject,
	}
}

type publisherImpl struct {
	PublisherMsg
	subject string
	header  nats.Header
}

func (p *publisherImpl) Subject() string {
	return p.subject
}

func (p *publisherImpl) WithSubject(subject string) Publisher {
	return &publisherImpl{p.PublisherMsg, subject, p.header}
}

func (p *publisherImpl) Header() *nats.Header {
	if p.header == nil {
		p.header = make(nats.Header)
	}
	return &p.header
}

func (p *publisherImpl) Publish(data []byte) error {
	if p.subject == "" {
		return errors.New("reply subject is not set")
	}
	if p.header == nil {
		p.header = make(nats.Header)
	}
	msg := nats.Msg{
		Subject: p.subject,
		Header:  p.header,
		Data:    data,
	}
	return p.PublishMsg(&msg)
}
