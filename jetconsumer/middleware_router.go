package jetconsumer

import (
	"context"
	"regexp"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

type Route interface {
	Match(string) bool
	Handler() Handler
}

type matcherFunc func(subject string) bool

type routeImpl struct {
	matcher matcherFunc
	handler Handler
}

func (r *routeImpl) Match(subj string) bool {
	return r.matcher(subj)
}

func (r *routeImpl) Handler() Handler {
	return r.handler
}

func exactMatcher(subject string) matcherFunc {
	return func(s string) bool {
		return s == subject
	}
}

func patternMatcher(pat string) matcherFunc {
	exp := "^" + strings.ReplaceAll(strings.ReplaceAll(regexp.QuoteMeta(pat), "\\*", "[^.]+"), "\\>", ".*") + "$"
	if exp == pat {
		return exactMatcher(pat)
	}
	re := regexp.MustCompile(exp)
	return func(subject string) bool {
		return re.MatchString(subject)
	}
}

func RouteSubject(subject string, handler Handler) Route {
	return &routeImpl{
		matcher: patternMatcher(subject),
		handler: handler,
	}
}

func Router(routes ...Route) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			for _, route := range routes {
				if route.Match(msg.Subject()) {
					return route.Handler().Serve(ctx, msg)
				}
			}
			return next.Serve(ctx, msg)
		})
	}
}
