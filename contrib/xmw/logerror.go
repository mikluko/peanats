package xmw

import (
	"context"

	"github.com/mikluko/peanats/xlog"
	"github.com/mikluko/peanats/xmsg"
)

type errorLogMiddlewareParams struct {
	logger    xlog.Logger
	propagate bool
}

type ErrorLogMiddlewareOption func(*errorLogMiddlewareParams)

// WithErrorLogMiddlewareLogger is a middleware option that specifies the logger
// to use for logging errors.
func WithErrorLogMiddlewareLogger(l xlog.Logger) ErrorLogMiddlewareOption {
	return func(p *errorLogMiddlewareParams) {
		p.logger = l
	}
}

// WithErrorLogMiddlewarePropagate is a middleware option that specifies whether
// the error should be propagated to the next handler.
func WithErrorLogMiddlewarePropagate(v bool) ErrorLogMiddlewareOption {
	return func(p *errorLogMiddlewareParams) {
		p.propagate = v
	}
}

func ErrorLogMiddleware(opts ...ErrorLogMiddlewareOption) xmsg.MsgMiddleware {
	p := errorLogMiddlewareParams{
		logger: xlog.StdLogger{},
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h xmsg.MsgHandler) xmsg.MsgHandler {
		return xmsg.MsgHandlerFunc(func(ctx context.Context, m xmsg.Msg) error {
			err := h.HandleMsg(ctx, m)
			if err == nil {
				return nil
			}
			p.logger.Log(ctx, "error", "message", err.Error())
			if p.propagate {
				return err
			}
			return nil
		})
	}
}
