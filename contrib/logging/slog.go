package logging

import (
	"context"
	"log/slog"

	"github.com/mikluko/peanats"
	"go.opentelemetry.io/otel/trace"
)

func SlogLogger(log *slog.Logger, lvl slog.Level) peanats.Logger {
	return &slogLogger{log, lvl}
}

type slogLogger struct {
	log *slog.Logger
	lvl slog.Level
}

func (l *slogLogger) Log(ctx context.Context, message string, args ...any) {
	l.log.Log(ctx, l.lvl, message, args...)
}

func TraceContextSlogLogger(log *slog.Logger, lvl slog.Level) peanats.Logger {
	return &traceContextSlogLogger{log, lvl}
}

type traceContextSlogLogger struct {
	log *slog.Logger
	lvl slog.Level
}

func (l *traceContextSlogLogger) Log(ctx context.Context, message string, args ...any) {
	log := l.log
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		log = log.With(slog.Group("otel",
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		))
	}
	l.log.Log(ctx, l.lvl, message, args...)
}
