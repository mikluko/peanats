package jetconsumer

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

type BatchMiddleware func(handler BatchHandler) BatchHandler

func ChainBatchMiddleware(h BatchHandler, mws ...BatchMiddleware) BatchHandler {
	for i := 0; i < len(mws); i++ {
		h = mws[i](h)
	}
	return h
}

// WithBatchAccessLog returns a middleware that logs access using the provided slog.Logger.
func WithBatchAccessLog(logger *slog.Logger, opts ...AccessLogOption) BatchMiddleware {
	params := &accessLogParams{
		level:   slog.LevelInfo,
		message: "",
	}
	for _, opt := range opts {
		opt(params)
	}
	return func(next BatchHandler) BatchHandler {
		return BatchHandlerFunc(func(ctx context.Context, batch jetstream.MessageBatch) error {
			b := accessLogBatch{MessageBatch: batch, ch: make(chan jetstream.Msg)}
			var (
				log *slog.Logger
				seq uint
			)
			go func() {
				for msg := range b.Messages() {
					if seq == 0 {
						log = logger.With("subject", msg.Subject())
						log.Info("batch started")
					}
					log.Debug("batch message", "subject", msg.Subject(), "headers", msg.Headers(), "sequence", seq)
					b.ch <- msg
				}
				log.Info("batch finished", "messages", seq)
			}()
			return next.Serve(ctx, &b)
		})
	}
}

var _ jetstream.MessageBatch = (*accessLogBatch)(nil)

type accessLogBatch struct {
	jetstream.MessageBatch
	ch chan jetstream.Msg
}

func (b *accessLogBatch) Messages() <-chan jetstream.Msg {
	return b.ch
}

// WithBatchErrorLog returns a middleware that logs errors using the provided logger.
// If propagate is true, the error is propagated to the outer handler.
func WithBatchErrorLog(logger *slog.Logger, propagate bool) BatchMiddleware {
	return func(next BatchHandler) BatchHandler {
		return BatchHandlerFunc(func(ctx context.Context, batch jetstream.MessageBatch) error {
			err := next.Serve(ctx, batch)
			if err != nil {
				logger.ErrorContext(ctx, "error", "err", err)
				if propagate {
					return err
				}
			}
			return nil
		})
	}
}
