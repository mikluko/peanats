package transport

import (
	"context"
	"errors"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

// Unsubscriber represents something that can be unsubscribed.
type Unsubscriber interface {
	Unsubscribe() error
}

// Subscription represents a message subscription that can pull messages
// and be unsubscribed.
type Subscription interface {
	Unsubscriber
	NextMsg(ctx context.Context) (peanats.Msg, error)
}

// Conn provides a typed message abstraction over a raw NATS connection.
type Conn interface {
	Publish(ctx context.Context, msg peanats.Msg) error
	Request(ctx context.Context, msg peanats.Msg) (peanats.Msg, error)
	Subscribe(ctx context.Context, subj string, opts ...SubscribeOption) (Subscription, error)
	SubscribeChan(ctx context.Context, subj string, ch chan peanats.Msg, opts ...SubscribeChanOption) (Unsubscriber, error)
	SubscribeHandler(ctx context.Context, subj string, handler peanats.MsgHandler, opts ...SubscribeHandlerOption) (Unsubscriber, error)
	Drain() error
	Close()
}

// Wrap wraps a *nats.Conn (and optional connection error) into a transport.Conn.
func Wrap(conn *nats.Conn, errs ...error) (Conn, error) {
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return New(conn), nil
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

func upstreamHandler(ctx context.Context, msgh peanats.MsgHandler, errh peanats.ErrorHandler, subm peanats.Submitter) nats.MsgHandler {
	return func(msg *nats.Msg) {
		subm.Submit(func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			err := msgh.HandleMsg(ctx, peanats.NewMsg(msg))
			if err != nil {
				errh.HandleError(ctx, err)
			}
		})
	}
}

// New creates a Conn from an upstream NATS connection.
func New(conn upstream) Conn {
	return &connImpl{conn}
}

type connImpl struct {
	nc upstream
}

type replier interface {
	Reply() string
}

func (c *connImpl) Publish(_ context.Context, msg peanats.Msg) error {
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

func (c *connImpl) Request(ctx context.Context, msg peanats.Msg) (peanats.Msg, error) {
	res, err := c.nc.RequestMsgWithContext(ctx, &nats.Msg{
		Subject: msg.Subject(),
		Header:  nats.Header(msg.Header()),
		Data:    msg.Data(),
	})
	if err != nil {
		return nil, err
	}
	return peanats.NewMsg(res), nil
}

func (c *connImpl) Drain() error {
	return c.nc.Drain()
}

func (c *connImpl) Close() {
	c.nc.Close()
}

// SubscribeOption configures synchronous subscriptions.
type SubscribeOption func(*subscribeParams)

type subscribeParams struct {
	queue string
}

// SubscribeQueue sets the queue group for a subscription.
func SubscribeQueue(name string) SubscribeOption {
	return func(p *subscribeParams) {
		p.queue = name
	}
}

func (c *connImpl) Subscribe(_ context.Context, subj string, opts ...SubscribeOption) (Subscription, error) {
	p := subscribeParams{}
	for _, o := range opts {
		o(&p)
	}
	var (
		sub *nats.Subscription
		err error
	)
	if p.queue != "" {
		sub, err = c.nc.QueueSubscribeSync(subj, p.queue)
	} else {
		sub, err = c.nc.SubscribeSync(subj)
	}
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

// SubscribeHandlerOption configures handler-based subscriptions.
type SubscribeHandlerOption func(*subscribeHandlerParams)

type subscribeHandlerParams struct {
	queue string
	subm  peanats.Submitter
	errh  peanats.ErrorHandler
}

// SubscribeHandlerSubmitter sets the task submitter for handler subscriptions.
func SubscribeHandlerSubmitter(subm peanats.Submitter) SubscribeHandlerOption {
	return func(p *subscribeHandlerParams) {
		p.subm = subm
	}
}

// SubscribeHandlerErrorHandler sets the error handler for handler subscriptions.
func SubscribeHandlerErrorHandler(errh peanats.ErrorHandler) SubscribeHandlerOption {
	return func(p *subscribeHandlerParams) {
		p.errh = errh
	}
}

// SubscribeHandlerQueue sets the queue group for handler subscriptions.
func SubscribeHandlerQueue(name string) SubscribeHandlerOption {
	return func(p *subscribeHandlerParams) {
		p.queue = name
	}
}

func (c *connImpl) SubscribeHandler(ctx context.Context, subj string, h peanats.MsgHandler, opts ...SubscribeHandlerOption) (Unsubscriber, error) {
	p := subscribeHandlerParams{
		subm: peanats.DefaultSubmitter,
		errh: peanats.DefaultErrorHandler,
	}
	for _, o := range opts {
		o(&p)
	}
	if p.queue != "" {
		return c.nc.QueueSubscribe(subj, p.queue, upstreamHandler(ctx, h, p.errh, p.subm))
	} else {
		return c.nc.Subscribe(subj, upstreamHandler(ctx, h, p.errh, p.subm))
	}
}

func (c *connImpl) mirror(mch chan peanats.Msg) chan *nats.Msg {
	nch := make(chan *nats.Msg, cap(mch))
	go func() {
		defer func() {
			if recover() != nil {
				for range nch {
				}
			}
		}()
		for msg := range nch {
			mch <- peanats.NewMsg(msg)
		}
	}()
	return nch
}

// SubscribeChanOption configures channel-based subscriptions.
type SubscribeChanOption func(*subscribeChanParams)

type subscribeChanParams struct {
	queue string
}

// SubscribeChanQueue sets the queue group for channel subscriptions.
func SubscribeChanQueue(name string) SubscribeChanOption {
	return func(p *subscribeChanParams) {
		p.queue = name
	}
}

func (c *connImpl) SubscribeChan(_ context.Context, subj string, ch chan peanats.Msg, opts ...SubscribeChanOption) (Unsubscriber, error) {
	p := subscribeChanParams{}
	for _, o := range opts {
		o(&p)
	}
	var (
		sub *nats.Subscription
		err error
	)
	if p.queue != "" {
		sub, err = c.nc.ChanQueueSubscribe(subj, p.queue, c.mirror(ch))
	} else {
		sub, err = c.nc.ChanSubscribe(subj, c.mirror(ch))
	}
	if err != nil {
		return nil, err
	}
	return &subscriptionImpl{sub}, nil
}

type subscriptionImpl struct {
	sub *nats.Subscription
}

func (s *subscriptionImpl) NextMsg(ctx context.Context) (peanats.Msg, error) {
	msg, err := s.sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return peanats.NewMsg(msg), nil
}

func (s *subscriptionImpl) Unsubscribe() error {
	return s.sub.Unsubscribe()
}
