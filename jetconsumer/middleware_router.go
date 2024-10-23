package jetconsumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
)

type Route interface {
	peanats.Matcher
	Handler
}

type routeImpl struct {
	peanats.Matcher
	Handler
}

func Handle(subj string, handler Handler) Route {
	return &routeImpl{peanats.NewMatcher(subj), handler}
}

func Router(routes ...Route) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			for _, route := range routes {
				if route.Match(msg.Subject()) {
					return route.Serve(ctx, msg)
				}
			}
			return next.Serve(ctx, msg)
		})
	}
}
