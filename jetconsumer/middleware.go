package jetconsumer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type Middleware func(handler Handler) Handler

func ChainMiddleware(h Handler, mws ...Middleware) Handler {
	for i := 0; i < len(mws); i++ {
		h = mws[i](h)
	}
	return h
}

// WithAccessLog returns a middleware that logs access using the provided slog.Logger.
func WithAccessLog(logger *slog.Logger, opts ...AccessLogOption) Middleware {
	params := &accessLogParams{
		level:   slog.LevelInfo,
		message: "",
	}
	for _, opt := range opts {
		opt(params)
	}
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			log := logger.With("subject", msg.Subject())
			hdr := msg.Headers()
			if hdr != nil {
				log = log.With("headers", hdr)
			}
			t := time.Now()
			err := next.Serve(ctx, msg)
			if err != nil {
				return err
			}
			log.Log(ctx, params.level, params.message, "latency", time.Since(t))
			return nil
		})
	}
}

type accessLogParams struct {
	level   slog.Level
	message string
}

type AccessLogOption func(*accessLogParams)

func WithAccessLogLevel(level slog.Level) AccessLogOption {
	return func(params *accessLogParams) {
		params.level = level
	}
}

func WithAccessLogMessage(message string) AccessLogOption {
	return func(params *accessLogParams) {
		params.message = message
	}
}

// WithErrorLog returns a middleware that logs errors using the provided logger.
// If propagate is true, the error is propagated to the outer handler.
func WithErrorLog(logger *slog.Logger, propagate bool) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			err := next.Serve(ctx, msg)
			if err != nil {
				logger.ErrorContext(ctx, "error", "subject", msg.Subject(), "err", err)
				if propagate {
					return err
				}
			}
			return nil
		})
	}
}

// WithAck returns a middleware that ACKs the message after handling it successfully.
func WithAck() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			err := next.Serve(ctx, msg)
			if err != nil {
				return err
			}
			return msg.Ack()
		})
	}
}

// WithNak returns a middleware that NAKs the message if the handler returns an error.
func WithNak() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			serveErr := next.Serve(ctx, msg)
			if serveErr != nil {
				if nakErr := msg.Nak(); nakErr != nil {
					return fmt.Errorf("%w: %w", nakErr, serveErr)
				}
			}
			return serveErr
		})
	}
}

// WithAckOnArrival returns a middleware that ACKs the message immediately on arrival.
func WithAckOnArrival() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			err := msg.Ack()
			if err != nil {
				return err
			}
			return next.Serve(ctx, msg)
		})
	}
}
