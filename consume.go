package peanats

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

type ConsumeOption func(*consumeParams)

type consumeParams struct {
	subm Submitter
	errh ErrorHandler
	opts []jetstream.PullConsumeOpt
}

// ConsumeSubmitter sets the workload submitter.
func ConsumeSubmitter(subm Submitter) ConsumeOption {
	return func(p *consumeParams) {
		p.subm = subm
	}
}

// ConsumeJetstreamOption sets the Jetstream pull consumer options.
func ConsumeJetstreamOption(opt jetstream.PullConsumeOpt) ConsumeOption {
	return func(p *consumeParams) {
		p.opts = append(p.opts, opt)
	}
}

// ConsumeErrorHandler sets the error handler.
func ConsumeErrorHandler(errh ErrorHandler) ConsumeOption {
	return func(p *consumeParams) {
		p.errh = errh
	}
}

// Consume implements consumer side of producer/consumer pattern.
func Consume(ctx context.Context, c jetstream.Consumer, h MessageHandler, opts ...ConsumeOption) error {
	p := consumeParams{
		subm: JustGoSubmitter{},
		errh: PanicErrorHandler{},
	}
	for _, o := range opts {
		o(&p)
	}
	cc, err := c.Consume(func(msg jetstream.Msg) {
		p.subm.Submit(func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			err := h.Handle(ctx, &messageJetstreamImpl{msg})
			if err != nil {
				err = p.errh.HandleError(ctx, err)
				if err != nil {
					panic(err)
				}
			}
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
