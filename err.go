package peanats

import (
	"context"
)

type ErrorHandler interface {
	HandleError(context.Context, error)
}

var DefaultErrorHandler ErrorHandler = errorHandlerImpl{}

type errorHandlerImpl struct{}

func (errorHandlerImpl) HandleError(_ context.Context, err error) {
	panic(err)
}
