package slog

import (
	"context"
	"log/slog"

	"github.com/mikluko/peanats"
)

func Logger(log *slog.Logger, lvl slog.Level) peanats.Logger {
	return &loggerImpl{log, lvl}
}

type loggerImpl struct {
	log *slog.Logger
	lvl slog.Level
}

func (l *loggerImpl) Log(ctx context.Context, message string, args ...any) {
	l.log.Log(ctx, l.lvl, message, args...)
}
