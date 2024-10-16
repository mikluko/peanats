package jetconsumer

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

//
// Consumer interface and constructor

type Consumer interface {
	Handler() jetstream.MessageHandler
}

func NewConsumer(opts ...ConsumerOpt) Consumer {
	c := &consumerImpl{
		baseCtx: context.Background(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

//
// Consumer implementation
//

type consumerImpl struct {
	baseCtx      context.Context
	routes       map[string]MessageHandler
	defaultRoute MessageHandler
	middleware   []Middleware
	exec         executor
}

func (c *consumerImpl) invoke(h MessageHandler, msg jetstream.Msg) {
	c.exec(func() {
		ctx, cancel := context.WithCancel(c.baseCtx)
		defer cancel()
		err := h.Serve(ctx, msg)
		if err != nil {
			panic(err)
		}
	})
}

func (c *consumerImpl) Handler() jetstream.MessageHandler {
	var h MessageHandler = MessageHandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
		subj := msg.Subject()
		route, found := c.routes[subj]
		if !found {
			if c.defaultRoute == nil {
				return fmt.Errorf("route not registered: %s", subj)
			}
			return c.defaultRoute.Serve(ctx, msg)
		}
		return route.Serve(ctx, msg)
	})
	for i := range c.middleware {
		h = c.middleware[i](h)
	}
	return func(msg jetstream.Msg) {
		c.invoke(h, msg)
	}
}

//
// Consumer options
//

type ConsumerOpt func(*consumerImpl)

func WithRoute[T any](subject string, h Handler[T]) ConsumerOpt {
	return func(impl *consumerImpl) {
		if impl.routes == nil {
			impl.routes = make(map[string]MessageHandler)
		}
		impl.routes[subject] = messageHandlerFrom(h)
	}
}

type executor func(func())

func WithExecutor(exec executor) ConsumerOpt {
	return func(impl *consumerImpl) {
		impl.exec = exec
	}
}

func WithDefaultRoute(h MessageHandler) ConsumerOpt {
	return func(impl *consumerImpl) {
		impl.defaultRoute = h
	}
}

func WithMiddleware(m ...Middleware) ConsumerOpt {
	return func(impl *consumerImpl) {
		impl.middleware = append(impl.middleware, m...)
	}
}
