package peanats

import (
	"context"
	"log"
)

type Logger interface {
	Log(ctx context.Context, message string, args ...any)
}

var DefaultLogger = loggerImpl{}

type loggerImpl struct{}

func (loggerImpl) Log(_ context.Context, message string, args ...any) {
	log.Println(message, args)
}
