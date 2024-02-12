package stream

import (
	"context"
	"errors"
	"fmt"

	"github.com/mikluko/newopt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"

	"github.com/mikluko/peanats"
)

type Client interface {
	Start(ctx context.Context, subj string, data []byte) (Receiver, error)
}

func NewClient(conn peanats.Connection, opts ...ClientOption) Client {
	defaults := []ClientOption{
		withConn(conn),
		WithReplySubjecter(ReplySubjectNUID()),
	}
	return newopt.NewP[clientImpl](append(defaults, opts...)...)
}

type ClientOption = newopt.Option[*clientImpl]

func withConn(conn peanats.Connection) ClientOption {
	return func(c *clientImpl) *clientImpl {
		c.conn = conn
		return c
	}
}

func WithReplySubjecter(rs ReplySubjecter) ClientOption {
	return func(c *clientImpl) *clientImpl {
		c.rs = rs
		return c
	}
}

var (
	ErrClient            = errors.New("stream client error")
	ErrProtocolViolation = fmt.Errorf("%w: %s", ErrClient, "protocol violation")
)

type clientImpl struct {
	conn peanats.Connection
	rs   ReplySubjecter
}

func (c *clientImpl) Start(ctx context.Context, subj string, data []byte) (Receiver, error) {
	uid := nuid.Next()
	msg := nats.NewMsg(subj)
	msg.Reply = c.rs.ReplySubject()
	msg.Data = data
	msg.Header = nats.Header{
		HeaderUID: []string{uid},
	}
	sub, err := c.conn.Subscribe(msg.Reply)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClient, "reply subject subscription failed")
	}
	err = c.conn.PublishMsg(msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClient, "initial message publish failed")
	}
	msg, err = sub.NextMsg(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrClient, "initial acknowledgment receive failed")
	}
	if msg.Header.Get(HeaderUID) == "" {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "UID header is absent in acknowledgment")
	}
	if msg.Header.Get(HeaderUID) != uid {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "unexpected UID in acknowledgment")
	}
	if msg.Header.Get(HeaderControl) == "" {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "control header is absent in acknowledgment")
	}
	if msg.Header.Get(HeaderControl) != HeaderControlAck {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "unexpected control header in acknowledgment")
	}
	if len(msg.Data) != 0 {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "unexpected data in acknowledgment")
	}
	return &receiverImpl{sub: sub, uid: uid}, nil
}

type TypedClient[A any, R any] interface {
	Start(ctx context.Context, subj string, arg *A) (TypedReceiver[R], error)
}
