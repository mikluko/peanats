package peanats

import "context"

type ErrorHandler interface {
	HandleError(context.Context, error) error
}

type PanicErrorHandler struct{}

func (PanicErrorHandler) HandleError(_ context.Context, err error) error {
	panic(err)
}

type LogErrorHandler struct {
	Logger Logger
}

func (l *LogErrorHandler) HandleError(ctx context.Context, err error) error {
	l.Logger.Log(ctx, "error", "message", err.Error())
	return nil
}
