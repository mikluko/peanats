package peanats

import (
	"context"

	"github.com/nats-io/nats.go"
)

type SubscribeOption func(*subscribeParams)

type subscribeParams struct {
	buffer uint
	queue  string
	subm   Submitter
	errh   ErrorHandler
}

// SubscribeBuffer sets the buffer size for the channel.
func SubscribeBuffer(size uint) SubscribeOption {
	return func(p *subscribeParams) {
		p.buffer = size
	}
}

// SubscribeQueue sets the queue name for the subscription. Please make sure you know what term queue
// means in the context for NATS subscription.
func SubscribeQueue(name string) SubscribeOption {
	return func(p *subscribeParams) {
		p.queue = name
	}
}

// SubscribeSubmitter sets the function that will be used to execute the handler.
func SubscribeSubmitter(subm Submitter) SubscribeOption {
	return func(p *subscribeParams) {
		p.subm = subm
	}
}

// SubscribeErrorHandler sets error handler for the subscription.
func SubscribeErrorHandler(errh ErrorHandler) SubscribeOption {
	return func(p *subscribeParams) {
		p.errh = errh
	}
}

// SubscribeChan returns a channel that will receive and handle messages from NATS subscription.
func SubscribeChan(ctx context.Context, h MessageHandler, opts ...SubscribeOption) (chan *nats.Msg, error) {
	p := subscribeParams{
		buffer: 1,
		subm:   JustGoSubmitter{},
		errh:   PanicErrorHandler{},
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
				p.subm.Submit(func() {
					err := h.Handle(ctx, &messageImpl{msg})
					if err != nil {
						err = p.errh.HandleError(ctx, err)
						if err != nil {
							panic(err)
						}
					}
				})
			}
		}
	}(ctx)
	return ch, nil
}
