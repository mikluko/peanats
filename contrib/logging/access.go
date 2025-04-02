package logging

import (
	"context"
	"time"

	"github.com/mikluko/peanats"
)

type accessLogMiddlewareParams struct {
	logger peanats.Logger
}

type AccessLogMiddlewareOption func(*accessLogMiddlewareParams)

// AccessLogMiddlewareLogger is a middleware option that specifies the logger
// to use for logging access.
func AccessLogMiddlewareLogger(l peanats.Logger) AccessLogMiddlewareOption {
	return func(p *accessLogMiddlewareParams) {
		p.logger = l
	}
}

// AccessLogMiddleware is a middleware that logs the message subject and headers
func AccessLogMiddleware(opts ...AccessLogMiddlewareOption) peanats.MsgMiddleware {
	p := accessLogMiddlewareParams{
		logger: peanats.DefaultLogger,
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
			t := time.Now()
			err := h.HandleMsg(ctx, m)
			if err != nil {
				return err
			}
			p.logger.Log(ctx, "done", "subject", m.Subject(), "header", m.Header(), "latency", time.Since(t))
			return nil
		})
	}
}
