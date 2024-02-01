package peanats

import (
	"errors"
	"sync"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	PublisherMsg
	Header() *nats.Header
	Publish(data []byte) error
}

type publisher struct {
	PublisherMsg
	subject string
	header  nats.Header
	once    sync.Once
}

func (p *publisher) init() {
	if p.header == nil {
		p.header = make(nats.Header)
	}
}

func (p *publisher) Header() *nats.Header {
	p.once.Do(p.init)
	return &p.header
}

func (p *publisher) Publish(data []byte) error {
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
