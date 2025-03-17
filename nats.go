package peanats

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"
)

func WrapConnection(conn *nats.Conn, errs ...error) (Connection, error) {
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return NewConnection(conn), nil
}

type Publisher interface {
	Publish(ctx context.Context, msg Msg) error
}

type Requester interface {
	Request(ctx context.Context, msg Msg) (Msg, error)
}

type Unsubscriber interface {
	Unsubscribe() error
}

type Subscription interface {
	Unsubscriber
	NextMsg(ctx context.Context) (Msg, error)
}

type Subscriber interface {
	Subscribe(ctx context.Context, subj string, opts ...SubscribeOption) (Subscription, error)
	SubscribeChan(ctx context.Context, subj string, ch chan Msg, opts ...SubscribeChanOption) (Unsubscriber, error)
	SubscribeHandler(ctx context.Context, subj string, handler MsgHandler, opts ...SubscribeHandlerOption) (Unsubscriber, error)
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
	Drainer
	Closer
}

type upstream interface {
	PublishMsg(msg *nats.Msg) error
	RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
	SubscribeSync(subj string) (*nats.Subscription, error)
	Subscribe(subj string, handler nats.MsgHandler) (*nats.Subscription, error)
	QueueSubscribeSync(subj, queue string) (*nats.Subscription, error)
	QueueSubscribe(subj, queue string, handler nats.MsgHandler) (*nats.Subscription, error)
	ChanSubscribe(subj string, ch chan *nats.Msg) (*nats.Subscription, error)
	ChanQueueSubscribe(subj, queue string, ch chan *nats.Msg) (*nats.Subscription, error)
	Drain() error
	Close()
}

var _ upstream = (*nats.Conn)(nil)

func upstreamHandler(ctx context.Context, msgh MsgHandler, errh ErrorHandler, subm Submitter) nats.MsgHandler {
	return func(msg *nats.Msg) {
		subm.Submit(func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			err := msgh.HandleMsg(ctx, NewMsg(msg))
			if err != nil {
				errh.HandleError(ctx, err)
			}
		})
	}
}

func NewConnection(conn upstream) Connection {
	return &connectionImpl{conn}
}

type connectionImpl struct {
	nc upstream
}

type replier interface {
	Reply() string
}

func (c *connectionImpl) Publish(ctx context.Context, msg Msg) error {
	m := nats.Msg{
		Subject: msg.Subject(),
		Header:  nats.Header(msg.Header()),
		Data:    msg.Data(),
	}
	if r, ok := msg.(replier); ok {
		m.Reply = r.Reply()
	}
	return c.nc.PublishMsg(&m)
}

func (c *connectionImpl) Request(ctx context.Context, msg Msg) (Msg, error) {
	res, err := c.nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: msg.Subject(),
		Header:  nats.Header(msg.Header()),
		Data:    msg.Data(),
	})
	if err != nil {
		return nil, err
	}
	return NewMsg(res), nil
}

func (c *connectionImpl) Drain() error {
	return c.nc.Drain()
}

func (c *connectionImpl) Close() {
	c.nc.Close()
}

type SubscribeOption func(*subscribeParams)

type subscribeParams struct {
	queue string
}

func SubscribeQueue(name string) SubscribeOption {
	return func(p *subscribeParams) {
		p.queue = name
	}
}

func (c *connectionImpl) Subscribe(ctx context.Context, subj string, opts ...SubscribeOption) (Subscription, error) {
	p := subscribeParams{}
	for _, o := range opts {
		o(&p)
	}
	var (
		sub *nats.Subscription
		err error
	)
	if p.queue != "" {
		sub, err = c.nc.SubscribeSync(subj)
	} else {
		sub, err = c.nc.QueueSubscribeSync(subj, p.queue)
	}
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

type SubscribeHandlerOption func(*subscribeHandlerParams)

type subscribeHandlerParams struct {
	queue string
	subm  Submitter
	errh  ErrorHandler
}

func SubscribeHandlerSubmitter(subm Submitter) SubscribeHandlerOption {
	return func(p *subscribeHandlerParams) {
		p.subm = subm
	}
}

func SubscribeHandlerErrorHandler(errh ErrorHandler) SubscribeHandlerOption {
	return func(p *subscribeHandlerParams) {
		p.errh = errh
	}
}

func (c *connectionImpl) SubscribeHandler(ctx context.Context, subj string, h MsgHandler, opts ...SubscribeHandlerOption) (Unsubscriber, error) {
	p := subscribeHandlerParams{
		subm: DefaultSubmitter,
		errh: DefaultErrorHandler,
	}
	for _, o := range opts {
		o(&p)
	}
	if p.queue != "" {
		return c.nc.Subscribe(subj, upstreamHandler(ctx, h, p.errh, p.subm))
	} else {
		return c.nc.QueueSubscribe(subj, p.queue, upstreamHandler(ctx, h, p.errh, p.subm))
	}
}

func (c *connectionImpl) mirror(mch chan Msg) chan *nats.Msg {
	nch := make(chan *nats.Msg, cap(mch))
	go func() {
		for msg := range nch {
			mch <- NewMsg(msg)
		}
	}()
	return nch
}

type SubscribeChanOption func(*subscribeChanParams)

type subscribeChanParams struct {
	queue string
}

func SubscribeChanQueue(name string) SubscribeChanOption {
	return func(p *subscribeChanParams) {
		p.queue = name
	}
}

func (c *connectionImpl) SubscribeChan(_ context.Context, subj string, ch chan Msg, opts ...SubscribeChanOption) (Unsubscriber, error) {
	p := subscribeChanParams{}
	for _, o := range opts {
		o(&p)
	}
	var (
		sub *nats.Subscription
		err error
	)
	if p.queue != "" {
		sub, err = c.nc.ChanSubscribe(subj, c.mirror(ch))
	} else {
		sub, err = c.nc.ChanQueueSubscribe(subj, p.queue, c.mirror(ch))
	}
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

type subscriptionImpl struct {
	sub *nats.Subscription
}

func (s *subscriptionImpl) NextMsg(ctx context.Context) (Msg, error) {
	msg, err := s.sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return NewMsg(msg), nil
}

func (s *subscriptionImpl) Unsubscribe() error {
	return s.sub.Unsubscribe()
}
