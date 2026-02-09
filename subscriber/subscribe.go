package subscriber

import (
	"context"

	"github.com/mikluko/peanats"
)

type SubscribeOption func(*subscribeParams)

type subscribeParams struct {
	buffer    uint
	queue     string
	disp      peanats.Dispatcher
	emptyDone bool
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

// SubscribeDispatcher sets the dispatcher for the subscription.
func SubscribeDispatcher(disp peanats.Dispatcher) SubscribeOption {
	return func(p *subscribeParams) {
		p.disp = disp
	}
}

// SubscribeEmptyDone modifies the behavior so that the channel will be closed upon
// receiving a message with an empty payload.
func SubscribeEmptyDone(x bool) SubscribeOption {
	return func(p *subscribeParams) {
		p.emptyDone = x
	}
}

// SubscribeChan returns a channel that will receive and handle messages from NATS subscription.
func SubscribeChan(ctx context.Context, h peanats.MsgHandler, opts ...SubscribeOption) (chan peanats.Msg, error) {
	p := subscribeParams{
		buffer:    1,
		disp:      peanats.DefaultDispatcher,
		emptyDone: false,
	}
	for _, o := range opts {
		o(&p)
	}
	ch := make(chan peanats.Msg, p.buffer)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg == nil {
					return
				}
				p.disp.Dispatch(func() error {
					return h.HandleMsg(ctx, msg)
				})
			}
		}
	}(ctx)
	return ch, nil
}
