package peanats

import "github.com/nats-io/nats.go"

type MsgPublisher interface {
	PublishMsg(msg *nats.Msg) error
}

type Connection interface {
	MsgPublisher
	ChanSubscribe(subj string, ch chan *nats.Msg) (Subscription, error)
	ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg) (Subscription, error)
	Drain() error
}

type Subscription interface {
	Unsubscribe() error
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

func (c *connectionImpl) Drain() error {
	return c.nc.Drain()
}

func (c *connectionImpl) ChanSubscribe(subj string, ch chan *nats.Msg) (Subscription, error) {
	sub, err := c.nc.ChanSubscribe(subj, ch)
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

func (c *connectionImpl) ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg) (Subscription, error) {
	sub, err := c.nc.ChanQueueSubscribe(subj, queue, ch)
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

func (c *connectionImpl) QueueSubscribe(subj, queue string) (*nats.Subscription, error) {
	return c.nc.QueueSubscribeSync(subj, queue)
}

type subscriptionImpl struct {
	sub *nats.Subscription
}

func (s *subscriptionImpl) Unsubscribe() error {
	return s.sub.Unsubscribe()
}
