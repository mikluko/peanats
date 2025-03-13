package logging

import (
	"context"

	"github.com/mikluko/peanats"
)

type errorLogMiddlewareParams struct {
	logger    peanats.Logger
	propagate bool
}

type ErrorLogMiddlewareOption func(*errorLogMiddlewareParams)

// ErrorLogMiddlewareLogger is a middleware option that specifies the logger
// to use for logging errors.
func ErrorLogMiddlewareLogger(l peanats.Logger) ErrorLogMiddlewareOption {
	return func(p *errorLogMiddlewareParams) {
		p.logger = l
	}
}

// ErrorLogMiddlewarePropagate is a middleware option that specifies whether
// the error should be propagated to the next handler.
func ErrorLogMiddlewarePropagate(v bool) ErrorLogMiddlewareOption {
	return func(p *errorLogMiddlewareParams) {
		p.propagate = v
	}
}

func ErrorLogMiddleware(opts ...ErrorLogMiddlewareOption) peanats.MsgMiddleware {
	p := errorLogMiddlewareParams{
		logger: peanats.DefaultLogger,
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h peanats.MsgHandler) peanats.MsgHandler {
		return peanats.MsgHandlerFunc(func(ctx context.Context, m peanats.Msg) error {
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
