package peanats

import (
	"context"
)

type ErrorHandler interface {
	HandleError(context.Context, error)
}

type ErrorHandlerFunc func(context.Context, error)

func (f ErrorHandlerFunc) HandleError(ctx context.Context, err error) {
	f(ctx, err)
}

var DefaultErrorHandler ErrorHandler = errorHandlerImpl{}

type errorHandlerImpl struct{}

func (errorHandlerImpl) HandleError(_ context.Context, err error) {
	panic(err)
}
