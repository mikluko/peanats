package jetconsumer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type Middleware func(handler MessageHandler) MessageHandler

// WithAccessLog returns a middleware that logs access using the provided slog.Logger.
func WithAccessLog(logger *slog.Logger, lvl slog.Level) Middleware {
	return func(next MessageHandler) MessageHandler {
		return MessageHandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
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
			log.Log(ctx, lvl, "", "latency", time.Since(t))
			return nil
		})
	}
}

// WithErrorLog returns a middleware that logs errors using the provided logger.
// If propagate is true, the error is propagated to the outer handler.
func WithErrorLog(logger *slog.Logger, propagate bool) Middleware {
	return func(next MessageHandler) MessageHandler {
		return MessageHandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			err := next.Serve(ctx, msg)
			if err != nil {
				logger.ErrorContext(ctx, "error", err)
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
	return func(next MessageHandler) MessageHandler {
		return MessageHandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			err := next.Serve(ctx, msg)
			if err != nil {
				return err
			}
			return msg.Ack()
		})
	}
}

// WithNak returns a middleware that NAKs the message if the handler returns an error.
// If propagate is true, the error is propagated to the outer handler.
func WithNak(propagate bool) Middleware {
	return func(next MessageHandler) MessageHandler {
		return MessageHandlerFunc(func(ctx context.Context, msg jetstream.Msg) error {
			serveErr := next.Serve(ctx, msg)
			if serveErr != nil {
				if nakErr := msg.Nak(); nakErr != nil {
					return fmt.Errorf("%w: %w", nakErr, serveErr)
				}
			}
			if propagate {
				return serveErr
			}
			return nil
		})
	}
}
