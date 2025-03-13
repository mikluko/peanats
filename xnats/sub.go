package xnats

import (
	"context"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats/xmsg"
)

type Unsubscriber interface {
	Unsubscribe() error
}

type Subscription interface {
	Unsubscriber
	NextMsg(ctx context.Context) (xmsg.Msg, error)
}

type subscriptionImpl struct {
	sub *nats.Subscription
}

func (s *subscriptionImpl) NextMsg(ctx context.Context) (xmsg.Msg, error) {
	msg, err := s.sub.NextMsgWithContext(ctx)
	if err != nil {
		return nil, err
	}
	return xmsg.New(msg), nil
}

func (s *subscriptionImpl) Unsubscribe() error {
	return s.sub.Unsubscribe()
}
