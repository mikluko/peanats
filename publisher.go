package peanats

import (
	"sync"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Header() *nats.Header
	Publish(data []byte) error
}

type publisher struct {
	header nats.Header
	conn   *nats.Conn
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

func (p *publisher) Publish(_ []byte) error {
	p.once.Do(p.init)
	return nil
}

type subjectPublisher struct {
	Publisher
	conn    *nats.Conn
	subject string
}

func (p *subjectPublisher) Publish(data []byte) error {
	msg := nats.Msg{
		Subject: p.subject,
		Header:  *p.Header(),
		Data:    data,
	}
	msg.Header = *p.Header()
	msg.Subject = p.subject
	msg.Data = data
	err := p.conn.PublishMsg(&msg)
	if err != nil {
		return err
	}
	return p.Publisher.Publish(data)
}
