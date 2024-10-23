package jetconsumer

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type JetstreamConsumer interface {
	Consume(jetstream.MessageHandler, ...jetstream.PullConsumeOpt) (jetstream.ConsumeContext, error)
}

type Consumer struct {
	BaseContext context.Context
	Consumer    JetstreamConsumer
	Handler     Handler
	Executor    func(func())

	cc jetstream.ConsumeContext
}

func (c *Consumer) Start() (err error) {
	if c.Consumer == nil {
		panic("jetstream consumer is not set")
	}
	if c.Handler == nil {
		panic("handler is not set")
	}
	if c.BaseContext == nil {
		c.BaseContext = context.Background()
	}
	if c.Executor == nil {
		c.Executor = func(f func()) { f() }
	}
	c.cc, err = c.Consumer.Consume(c.handler(), jetstream.PullExpiry(time.Second*10))
	if err != nil {
		return err
	}
	return nil
}

func (c *Consumer) Stop() {
	c.cc.Stop()
}

func (c *Consumer) handler() jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		c.Executor(func() {
			ctx, cancel := context.WithCancel(c.BaseContext)
			defer cancel()
			err := c.Handler.Serve(ctx, msg)
			if err != nil {
				panic(err)
			}
		})
	}
}
