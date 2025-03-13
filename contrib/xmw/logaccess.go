package xmw

import (
	"context"

	"github.com/mikluko/peanats/xlog"
	"github.com/mikluko/peanats/xmsg"
)

type accessLogMiddlewareParams struct {
	logger xlog.Logger
}

type AccessLogMiddlewareOption func(*accessLogMiddlewareParams)

// WithAccessLogMiddlewareLogger is a middleware option that specifies the logger
// to use for logging access.
func WithAccessLogMiddlewareLogger(l xlog.Logger) AccessLogMiddlewareOption {
	return func(p *accessLogMiddlewareParams) {
		p.logger = l
	}
}

// AccessLogMiddleware is a middleware that logs the message subject and headers
func AccessLogMiddleware(opts ...AccessLogMiddlewareOption) xmsg.MsgMiddleware {
	p := accessLogMiddlewareParams{
		logger: xlog.StdLogger{},
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h xmsg.MsgHandler) xmsg.MsgHandler {
		return xmsg.MsgHandlerFunc(func(ctx context.Context, m xmsg.Msg) error {
			err := h.HandleMsg(ctx, m)
			if err != nil {
				return err
			}
			p.logger.Log(ctx, "done", "subject", m.Subject(), "header", m.Header())
			return nil
		})
	}
}
