package peanats

import (
	"sync"

	"github.com/nats-io/nats.go"
)

type Acker interface {
	Ack(data []byte) error
}

type Publisher interface {
	Header() *nats.Header
	Publish(data []byte) error
}

type AckPublisher interface {
	Acker
	Publisher
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

func (p *publisher) Ack(data []byte) error {
	p.once.Do(p.init)
	if p.msg == nil || p.msg.Sub == nil {
		return nil
	}
	if p.msg.Reply == "" {
		return nil
	}
	res := new(nats.Msg)
	res.Data = data
	res.Header = p.header
	return p.msg.RespondMsg(res)
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

func (p *subjectPublisher) Ack(data []byte) error {
	if ack, ok := p.Publisher.(Acker); ok {
		return ack.Ack(data)
	}
	return nil
}
