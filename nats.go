package peanats

import (
	"context"
	"github.com/nats-io/nats.go"
)

type PublisherMsg interface {
	PublishMsg(msg *nats.Msg) error
}

type RequesterMsg interface {
	RequestMsg(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
}

type Subscriber interface {
	Subscribe(subj string) (Subscription, error)
	QueueSubscribe(subj, queue string) (Subscription, error)
}

type ChanSubscriber interface {
	ChanSubscribe(subj string, ch chan *nats.Msg) (Unsubscriber, error)
	ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg) (Unsubscriber, error)
}

type Connection interface {
	PublisherMsg
	RequesterMsg
	Subscriber
	ChanSubscriber
	Drain() error
}

type Unsubscriber interface {
	Unsubscribe() error
}

type Subscription interface {
	Unsubscriber
	NextMsg(ctx context.Context) (*nats.Msg, error)
}

func NATS(nc *nats.Conn) Connection {
	return &connectionImpl{nc}
}

type connectionImpl struct {
	nc *nats.Conn
}

func (c *connectionImpl) PublishMsg(msg *nats.Msg) error {
	return c.nc.PublishMsg(msg)
}

func (c *connectionImpl) RequestMsg(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	return c.nc.RequestMsgWithContext(ctx, msg)
}

func (c *connectionImpl) Drain() error {
	return c.nc.Drain()
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

func (c *connectionImpl) ChanSubscribe(subj string, ch chan *nats.Msg) (Unsubscriber, error) {
	sub, err := c.nc.ChanSubscribe(subj, ch)
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

func (c *connectionImpl) ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg) (Unsubscriber, error) {
	sub, err := c.nc.ChanQueueSubscribe(subj, queue, ch)
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

type subscriptionImpl struct {
	sub *nats.Subscription
}

func (s *subscriptionImpl) NextMsg(ctx context.Context) (*nats.Msg, error) {
	return s.sub.NextMsgWithContext(ctx)
}

func (s *subscriptionImpl) Unsubscribe() error {
	return s.sub.Unsubscribe()
}
