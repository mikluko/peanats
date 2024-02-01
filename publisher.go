package peanats

import (
	"errors"
	"sync"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Header() *nats.Header
	Publish(data []byte) error
}

type publisher struct {
	header nats.Header
	msgpub MsgPublisher
	msg    *nats.Msg
	once   sync.Once
}

func (p *publisher) init() {
	p.header = make(nats.Header)
}

func (p *publisher) Header() *nats.Header {
	p.once.Do(p.init)
	return &p.header
}

func (p *publisher) Publish(data []byte) error {
	p.once.Do(p.init)
	if p.msg.Reply == "" {
		return errors.New("reply subject is not set")
	}
	msg := nats.Msg{
		Subject: p.msg.Reply,
		Header:  p.header,
		Data:    data,
	}
	return p.msgpub.PublishMsg(&msg)
}

type subjectPublisher struct {
	Publisher
	msgpub  MsgPublisher
	subject string
}

func (p *subjectPublisher) Publish(data []byte) error {
	msg := nats.Msg{
		Subject: p.subject,
		Header:  *p.Header(),
		Data:    data,
	}
	return p.msgpub.PublishMsg(&msg)
}
