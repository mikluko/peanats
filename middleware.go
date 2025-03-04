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

type AccessLogger interface {
	Log(ctx context.Context, message string, args ...interface{})
}

func AccessLogMiddleware(log AccessLogger) Middleware {
	return func(h Handler) Handler {
		return HandlerFunc(func(ctx context.Context, d Dispatcher, m Message) {
			h.Handle(ctx, d, m)
			log.Log(ctx, "message processed", "subject", m.Subject(), "header", m.Header())
		})
	}
}

func NewSlogAccessLogger(l *slog.Logger) AccessLogger {
	return &slogAccessLogger{l}
}

type slogAccessLogger struct {
	l *slog.Logger
}

func (l *slogAccessLogger) Log(ctx context.Context, message string, args ...interface{}) {
	l.l.InfoContext(ctx, message, args...)
}
