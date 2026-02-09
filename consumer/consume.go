package consumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
)

type ConsumeOption func(*consumeParams)

type consumeParams struct {
	disp peanats.Dispatcher
	opts []jetstream.PullConsumeOpt
}

// ConsumeDispatcher sets the dispatcher.
func ConsumeDispatcher(disp peanats.Dispatcher) ConsumeOption {
	return func(p *consumeParams) {
		p.disp = disp
	}
}

// ConsumeJetstreamOption sets the Jetstream pull consumer options.
func ConsumeJetstreamOption(opt jetstream.PullConsumeOpt) ConsumeOption {
	return func(p *consumeParams) {
		p.opts = append(p.opts, opt)
	}
}

// ConsumePullMaxMessages sets the maximum number of messages to pull per request.
func ConsumePullMaxMessages(size int) ConsumeOption {
	return func(p *consumeParams) {
		p.opts = append(p.opts, jetstream.PullMaxMessages(size))
	}
}

type consumer interface {
	Consume(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)
}

// Consume implements consumer side of producer/consumer pattern.
func Consume(ctx context.Context, c consumer, h peanats.MsgHandler, opts ...ConsumeOption) error {
	p := consumeParams{
		disp: peanats.DefaultDispatcher,
		opts: []jetstream.PullConsumeOpt{},
	}
	for _, o := range opts {
		o(&p)
	}
	cc, err := c.Consume(func(m jetstream.Msg) {
		p.disp.Dispatch(func() error {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			return h.HandleMsg(ctx, peanats.NewJetstream(m))
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
