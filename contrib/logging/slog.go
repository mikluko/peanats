package logging

import (
	"context"
	"log/slog"
)

func SlogLogger(log *slog.Logger, lvl slog.Level) Logger {
	return &slogLogger{log, lvl}
}

type slogLogger struct {
	log *slog.Logger
	lvl slog.Level
}

func (l *slogLogger) Log(ctx context.Context, message string, args ...any) {
	l.log.Log(ctx, l.lvl, message, args...)
}
