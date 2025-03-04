package peanats

import (
	"context"
	"log/slog"
)

type Middleware func(Handler) Handler

func ChainMiddleware(h Handler, mw ...Middleware) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

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
func AccessLogMiddleware(opts ...AccessLogMiddlewareOption) Middleware {
	p := accessLogMiddlewareParams{
		logger: NewSlogLogger(slog.Default(), slog.LevelInfo),
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h Handler) Handler {
		return HandlerFunc(func(ctx context.Context, d Dispatcher, m Message) {
			h.Handle(ctx, d, m)
			p.logger.Log(ctx, "done", "subject", m.Subject(), "header", m.Header())
		})
	}
}

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

func ErrorLogMiddleware(opts ...ErrorLogMiddlewareOption) Middleware {
	p := errorLogMiddlewareParams{
		logger: NewSlogLogger(slog.Default(), slog.LevelError),
	}
	for _, o := range opts {
		o(&p)
	}
	return func(h Handler) Handler {
		return HandlerFunc(func(ctx context.Context, d Dispatcher, m Message) {
			h.Handle(ctx, errorLogDispatcher{d, p}, m)
		})
	}
}

type errorLogDispatcher struct {
	Dispatcher
	errorLogMiddlewareParams
}

func (d errorLogDispatcher) Error(ctx context.Context, err error) {
	d.logger.Log(ctx, "error", "message", err.Error())
	if d.propagate {
		d.Dispatcher.Error(ctx, err)
	}
}

type Logger interface {
	Log(ctx context.Context, message string, args ...interface{})
}

func NewSlogLogger(log *slog.Logger, lvl slog.Level) Logger {
	return &slogLogger{log, lvl}
}

type slogLogger struct {
	log *slog.Logger
	lvl slog.Level
}

func (l *slogLogger) Log(ctx context.Context, message string, args ...any) {
	l.log.Log(ctx, l.lvl, message, args...)
}
