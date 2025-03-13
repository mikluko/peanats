package xerr

import (
	"context"

	"github.com/mikluko/peanats/xlog"
)

type ErrorHandler interface {
	HandleError(context.Context, error)
}

type PanicErrorHandler struct{}

func (PanicErrorHandler) HandleError(_ context.Context, err error) {
	panic(err)
}

type LogErrorHandler struct {
	Logger xlog.Logger
}

func (l *LogErrorHandler) HandleError(ctx context.Context, err error) error {
	l.Logger.Log(ctx, "error", "message", err.Error())
	return nil
}
