package peaserve

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

// ServeChan returns a channel that will receive and handle messages from NATS subscription.
func ServeChan(ctx context.Context, mh peanats.Handler, opts ...Option) (chan *nats.Msg, error) {
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
					rq := peanats.NewRequestMessage(msg)
					header := peanats.Header{
						peanats.HeaderContentType: []string{msg.Header.Get(peanats.HeaderContentType)},
					}
					d := &messageDispatcher{rq: rq, header: header}
					mh.Handle(ctx, d, rq)
				})
			}
		}
	}(ctx)
	return ch, nil
}

func Handler[RQ any, RS any](h peanats.ArgumentHandler[RQ]) peanats.Handler {
	pool := xargpool.New[RQ]()
	return peanats.HandlerFunc(func(ctx context.Context, md peanats.Dispatcher, rq peanats.Message) {
		x := pool.Acquire(ctx)
		v := x.Value()
		defer x.Release()

		err := peanats.Unmarshal(rq.Header(), rq.Data(), v)
		if err != nil {
			md.Error(ctx, err)
			return
		}

		ad := &argumentDispatcher[RS]{md, rq.(peanats.RequestMessage)}
		h.Handle(ctx, ad, argumentImpl[RQ]{rq, v})
	})
}

type messageDispatcher struct {
	rq     peanats.RequestMessage
	header peanats.Header
}

func (d *messageDispatcher) Header() peanats.Header {
	return d.header
}

func (d *messageDispatcher) Error(_ context.Context, err error) {
	panic(err)
}

type Dispatcher[T any] interface {
	peanats.Dispatcher
	Respond(context.Context, *T) error
}

type argumentDispatcher[T any] struct {
	peanats.Dispatcher
	rq peanats.RequestMessage
}

func (d *argumentDispatcher[T]) Respond(ctx context.Context, res *T) error {
	p, err := peanats.Marshal(d.Header(), res)
	if err != nil {
		return err
	}
	return d.rq.Respond(ctx, responseMessage{d.Header(), p})
}

var _ Dispatcher[any] = (*argumentDispatcher[any])(nil)

type responseMessage struct {
	header peanats.Header
	data   []byte
}

func (m responseMessage) Subject() string {
	panic("programming error: this method should not be called on response message")
}

func (m responseMessage) Header() peanats.Header {
	return m.header
}

func (m responseMessage) Data() []byte {
	return m.data
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
