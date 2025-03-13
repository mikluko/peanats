package peaconsumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats/xerr"
	"github.com/mikluko/peanats/xmsg"
	"github.com/mikluko/peanats/xsubm"
)

type ConsumeOption func(*consumeParams)

type consumeParams struct {
	subm xsubm.Submitter
	errh xerr.ErrorHandler
	opts []jetstream.PullConsumeOpt
}

// ConsumeSubmitter sets the workload submitter.
func ConsumeSubmitter(subm xsubm.Submitter) ConsumeOption {
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
func ConsumeErrorHandler(errh xerr.ErrorHandler) ConsumeOption {
	return func(p *consumeParams) {
		p.errh = errh
	}
}

type consumer interface {
	Consume(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)
}

// Consume implements consumer side of producer/consumer pattern.
func Consume(ctx context.Context, c consumer, h xmsg.MsgHandler, opts ...ConsumeOption) error {
	p := consumeParams{
		subm: xsubm.JustGoSubmitter{},
		errh: xerr.PanicErrorHandler{},
		opts: []jetstream.PullConsumeOpt{},
	}
	for _, o := range opts {
		o(&p)
	}
	cc, err := c.Consume(func(m jetstream.Msg) {
		p.subm.Submit(func() {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			err := h.HandleMsg(ctx, xmsg.NewJetstream(m))
			if err != nil {
				p.errh.HandleError(ctx, err)
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
