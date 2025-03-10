package peaconsumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xargpool"
)

type Option func(*params)

type params struct {
	exec func(func())
	opts []jetstream.PullConsumeOpt
}

func defaults() params {
	return params{
		exec: func(f func()) { f() },
	}
}

// WithExecutor sets the executor.
// If not set, the default executor runs the handler synchronously in the same goroutine.
func WithExecutor(e func(func())) Option {
	return func(p *params) {
		p.exec = e
	}
}

// WithJetstreamOption sets the Jetstream pull consumer options.
func WithJetstreamOption(opt jetstream.PullConsumeOpt) Option {
	return func(p *params) {
		p.opts = append(p.opts, opt)
	}
}

type jetstreamConsumer interface {
	Consume(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)
}

// Consume implements consumer side of producer/consumer pattern.
func Consume(ctx context.Context, c jetstreamConsumer, h peanats.Handler, opts ...Option) error {
	p := defaults()
	for _, o := range opts {
		o(&p)
	}
	cc, err := c.Consume(func(msg jetstream.Msg) {
		p.exec(func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			m := peanats.NewJetstreamMessage(msg)
			d := dispatcherImpl{m}
			h.Handle(ctx, &d, m)
		})
	}, p.opts...)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		cc.Stop()
	}()
	return nil
}

func Handler[T any](h peanats.ArgumentHandler[T]) peanats.Handler {
	pool := xargpool.New[T]()
	return peanats.HandlerFunc(func(ctx context.Context, d peanats.Dispatcher, m peanats.Message) {
		x := pool.Acquire(ctx)
		v := x.Value()
		defer x.Release()

		err := peanats.Unmarshal(m.Header(), m.Data(), v)
		if err != nil {
			d.Error(ctx, err)
			return
		}

		msg := m.(peanats.JetstreamMessage) // panic if not JetstreamMessage is intended
		h.Handle(ctx, dispatcherImpl{msg}, argumentImpl[T]{msg, v})
	})
}

// Dispatcher interface defines a dispatcher that can acknowledge the
// completion of the processing of the message.
type Dispatcher interface {
	peanats.Dispatcher
	Ack(context.Context, ...peanats.AckOption) error
	Nak(context.Context, ...peanats.NakOption) error
	Term(context.Context, ...peanats.TermOption) error
	InProgress(ctx context.Context) error
}

var _ Dispatcher = (*dispatcherImpl)(nil)

type dispatcherImpl struct {
	msg peanats.JetstreamMessage
}

func (d dispatcherImpl) Header() peanats.Header {
	return d.msg.Header()
}

func (d dispatcherImpl) Error(_ context.Context, err error) {
	if err != nil {
		panic(err)
	}
}

func (d dispatcherImpl) Ack(ctx context.Context, opts ...peanats.AckOption) error {
	return d.msg.Ack(ctx, opts...)
}

func (d dispatcherImpl) Nak(ctx context.Context, opts ...peanats.NakOption) error {
	return d.msg.Nak(ctx, opts...)
}

func (d dispatcherImpl) Term(ctx context.Context, opts ...peanats.TermOption) error {
	return d.msg.Term(ctx, opts...)
}

func (d dispatcherImpl) InProgress(ctx context.Context) error {
	return d.msg.InProgress(ctx)
}

var _ peanats.Argument[any] = (*argumentImpl[any])(nil)

type argumentImpl[T any] struct {
	m peanats.Message
	v *T
}

func (a argumentImpl[T]) Subject() string {
	return a.m.Subject()
}

func (a argumentImpl[T]) Header() peanats.Header {
	return a.m.Header()
}

func (a argumentImpl[T]) Payload() *T {
	return a.v
}
