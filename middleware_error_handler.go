package peanats

import (
	"go.uber.org/multierr"
)

func ErrorHandlerMiddleware(next Handler) Handler {
	return HandlerFunc(func(pub Publisher, req Request) error {
		errPub := errorHandlerPublisher{
			Publisher: pub,
		}
		err1 := next.Serve(&errPub, req)
		if err1 != nil {
			err2 := PublishError(&errPub, err1)
			if err2 != nil {
				return multierr.Combine(err1, err2)
			}
		}
		return nil
	})
}

type errorHandlerPublisher struct {
	Publisher
}
