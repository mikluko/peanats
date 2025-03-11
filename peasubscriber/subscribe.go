package peasubscriber

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xargpool"
)

type Option func(*params)

type params struct {
	buffer uint
	queue  string
	exec   func(func())
}

// WithBuffer sets the buffer size for the channel.
func WithBuffer(size uint) Option {
	return func(p *params) {
		p.buffer = size
	}
}

// WithQueue sets the queue name for the subscription. Please make sure you know what term queue
// means in the context for NATS subscription.
func WithQueue(name string) Option {
	return func(p *params) {
		p.queue = name
	}
}

// WithExecutor sets the function that will be used to execute the handler.
func WithExecutor(e func(func())) Option {
	return func(p *params) {
		p.exec = e
	}
}

// SubscribeChan returns a channel that will receive and handle messages from NATS subscription.
func SubscribeChan(ctx context.Context, h peanats.Handler, opts ...Option) (chan *nats.Msg, error) {
	p := params{
		buffer: 1,
		exec:   func(f func()) { f() },
	}
	for _, o := range opts {
		o(&p)
	}
	ch := make(chan *nats.Msg, p.buffer)
	go func(ctx context.Context) {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				p.exec(func() {
					m := peanats.NewMessage(msg)
					d := dispatcherImpl{m, make(peanats.Header)}
					h.Handle(ctx, d, m)
				})
			}
		}
	}(ctx)
	return ch, nil
}

func Handler[T any](h peanats.ArgumentHandler[T]) peanats.Handler {
	pool := xargpool.New[T]()
	return peanats.HandlerFunc(func(ctx context.Context, d peanats.Dispatcher, m peanats.Message) {
		x := pool.Acquire(ctx)
		v := x.Value()
		defer x.Release()

		err := peanats.Unmarshal(peanats.ContentTypeHeader(m.Header()), m.Data(), v)
		if err != nil {
			d.Error(ctx, err)
			return
		}

		h.Handle(ctx, d, argumentImpl[T]{m, v})
	})
}

type dispatcherImpl struct {
	m peanats.Message
	h peanats.Header
}

func (d dispatcherImpl) Header() peanats.Header {
	return d.h
}

func (d dispatcherImpl) Error(_ context.Context, err error) {
	panic(err)
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
