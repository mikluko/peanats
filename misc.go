package peanats

import (
	"context"
	"log"
)

type Logger interface {
	Log(ctx context.Context, message string, args ...any)
}

type StdLogger struct{}

func (StdLogger) Log(_ context.Context, message string, args ...any) {
	log.Println(message, args)
}

type Submitter interface {
	Submit(func())
}

type JustGoSubmitter struct{}

func (JustGoSubmitter) Submit(f func()) {
	go f()
}
