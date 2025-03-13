package peasubscriber

import (
	"context"

	"github.com/mikluko/peanats/xerr"
	"github.com/mikluko/peanats/xmsg"
	"github.com/mikluko/peanats/xsubm"
)

type SubscribeOption func(*subscribeParams)

type subscribeParams struct {
	buffer    uint
	queue     string
	subm      xsubm.Submitter
	errh      xerr.ErrorHandler
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

// SubscribeSubmitter sets the function that will be used to execute the handler.
func SubscribeSubmitter(subm xsubm.Submitter) SubscribeOption {
	return func(p *subscribeParams) {
		p.subm = subm
	}
}

// SubscribeErrorHandler sets error handler for the subscription.
func SubscribeErrorHandler(errh xerr.ErrorHandler) SubscribeOption {
	return func(p *subscribeParams) {
		p.errh = errh
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
func SubscribeChan(ctx context.Context, h xmsg.MsgHandler, opts ...SubscribeOption) (chan xmsg.Msg, error) {
	p := subscribeParams{
		buffer:    1,
		subm:      xsubm.JustGoSubmitter{},
		errh:      xerr.PanicErrorHandler{},
		emptyDone: false,
	}
	for _, o := range opts {
		o(&p)
	}
	ch := make(chan xmsg.Msg, p.buffer)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg == nil {
					return
				}
				p.subm.Submit(func() {
					err := h.HandleMsg(ctx, msg)
					if err != nil {
						p.errh.HandleError(ctx, err)
					}
				})
			}
		}
	}(ctx)
	return ch, nil
}
