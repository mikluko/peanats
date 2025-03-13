package peanats

import (
	"context"
)

type accessLogMiddlewareParams struct {
	logger Logger
}

type AccessLogMiddlewareOption func(*accessLogMiddlewareParams)

// WithAccessLogMiddlewareLogger is a middleware option that specifies the logger
// to use for logging access.
func WithAccessLogMiddlewareLogger(l Logger) AccessLogMiddlewareOption {
	return func(p *accessLogMiddlewareParams) {
		p.logger = l
	}
}

// AccessLogMiddleware is a middleware that logs the message subject and headers
func AccessLogMiddleware(opts ...AccessLogMiddlewareOption) MessageMiddleware {
	p := accessLogMiddlewareParams{
		logger: StdLogger{},
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h MessageHandler) MessageHandler {
		return MessageHandlerFunc(func(ctx context.Context, m Message) error {
			err := h.Handle(ctx, m)
			if err != nil {
				return err
			}
			p.logger.Log(ctx, "done", "subject", m.Subject(), "header", m.Header())
			return nil
		})
	}
}
