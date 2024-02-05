package peanats

import (
	"errors"
	"sync"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	PublisherMsg
	Subject() string
	Header() *nats.Header
	Publish(data []byte) error
}

type publisherImpl struct {
	PublisherMsg
	subject string
	header  nats.Header
	once    sync.Once
}

func (p *publisherImpl) init() {
	if p.header == nil {
		p.header = make(nats.Header)
	}
}

func (p *publisherImpl) Subject() string {
	p.once.Do(p.init)
	return p.subject
}

func (p *publisherImpl) Header() *nats.Header {
	p.once.Do(p.init)
	return &p.header
}

func (p *publisherImpl) Publish(data []byte) error {
	p.once.Do(p.init)
	if p.subject == "" {
		return errors.New("reply subject is not set")
	}
	msg := nats.Msg{
		Subject: p.subject,
		Header:  p.header,
		Data:    data,
	}
	return p.PublishMsg(&msg)
}
