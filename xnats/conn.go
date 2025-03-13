package xnats

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats/xmsg"
)

type Publisher interface {
	Publish(ctx context.Context, msg xmsg.Msg) error
}

type Requester interface {
	Request(ctx context.Context, msg xmsg.Msg) (xmsg.Msg, error)
}

type Subscriber interface {
	Subscribe(subj string) (Subscription, error)
	QueueSubscribe(subj, queue string) (Subscription, error)
}

type ChanSubscriber interface {
	ChanSubscribe(subj string, ch chan xmsg.Msg) (Unsubscriber, error)
	ChanQueueSubscribe(subj, queue string, ch chan xmsg.Msg) (Unsubscriber, error)
}

type Drainer interface {
	Drain() error
}

type Closer interface {
	Close()
}

type Connection interface {
	Publisher
	Requester
	Subscriber
	ChanSubscriber
	Drainer
	Closer
}

type upstream interface {
	PublishMsg(msg *nats.Msg) error
	RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
	SubscribeSync(subj string) (*nats.Subscription, error)
	QueueSubscribeSync(subj, queue string) (*nats.Subscription, error)
	ChanSubscribe(subj string, ch chan *nats.Msg) (*nats.Subscription, error)
	ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg) (*nats.Subscription, error)
	Drain() error
	Close()
}

var _ upstream = (*nats.Conn)(nil)

func Wrap(conn *nats.Conn, errs ...error) (Connection, error) {
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return NewConnection(conn), nil
}

func NewConnection(conn upstream) Connection {
	return &connectionImpl{conn}
}

type connectionImpl struct {
	nc upstream
}

func (c *connectionImpl) Publish(ctx context.Context, msg xmsg.Msg) error {
	return c.nc.PublishMsg(&nats.Msg{
		Subject: msg.Subject(),
		Header:  nats.Header(msg.Header()),
		Data:    msg.Data(),
	})
}

func (c *connectionImpl) Request(ctx context.Context, msg xmsg.Msg) (xmsg.Msg, error) {
	res, err := c.nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: msg.Subject(),
		Header:  nats.Header(msg.Header()),
		Data:    msg.Data(),
	})
	if err != nil {
		return nil, err
	}
	return xmsg.New(res), nil
}

func (c *connectionImpl) Drain() error {
	return c.nc.Drain()
}

func (c *connectionImpl) Close() {
	c.nc.Close()
}

func (c *connectionImpl) Subscribe(subj string) (Subscription, error) {
	sub, err := c.nc.SubscribeSync(subj)
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

func (c *connectionImpl) QueueSubscribe(subj, queue string) (Subscription, error) {
	sub, err := c.nc.QueueSubscribeSync(subj, queue)
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

func (c *connectionImpl) mirror(mch chan xmsg.Msg) chan *nats.Msg {
	nch := make(chan *nats.Msg, cap(mch))
	go func() {
		for msg := range nch {
			mch <- xmsg.New(msg)
		}
	}()
	return nch
}

func (c *connectionImpl) ChanSubscribe(subj string, ch chan xmsg.Msg) (Unsubscriber, error) {
	sub, err := c.nc.ChanSubscribe(subj, c.mirror(ch))
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

func (c *connectionImpl) ChanQueueSubscribe(subj, queue string, ch chan xmsg.Msg) (Unsubscriber, error) {
	sub, err := c.nc.ChanQueueSubscribe(subj, queue, c.mirror(ch))
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}
