package peanats

import "context"

type errorLogMiddlewareParams struct {
	logger    Logger
	propagate bool
}

type ErrorLogMiddlewareOption func(*errorLogMiddlewareParams)

// WithErrorLogMiddlewareLogger is a middleware option that specifies the logger
// to use for logging errors.
func WithErrorLogMiddlewareLogger(l Logger) ErrorLogMiddlewareOption {
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

func ErrorLogMiddleware(opts ...ErrorLogMiddlewareOption) MessageMiddleware {
	p := errorLogMiddlewareParams{
		logger: StdLogger{},
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h MessageHandler) MessageHandler {
		return MessageHandlerFunc(func(ctx context.Context, m Message) error {
			err := h.Handle(ctx, m)
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
