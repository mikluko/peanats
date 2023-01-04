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
			var err2 error
			if errPub.isAcker() && !errPub.acked {
				err2 = AckError(&errPub, err1)
			} else {
				err2 = PublishError(&errPub, err1)
			}
			if err2 != nil {
				return multierr.Combine(err1, err2)
			}
		}
		return nil
	})
}

type errorHandlerPublisher struct {
	Publisher
	acked bool
}

func (p *errorHandlerPublisher) isAcker() bool {
	_, ok := p.Publisher.(Acker)
	return ok
}

func (p *errorHandlerPublisher) Ack(data []byte) error {
	ack, ok := p.Publisher.(Acker)
	if !ok {
		return nil
	}
	p.acked = true
	return ack.Ack(data)
}
