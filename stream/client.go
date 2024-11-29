package stream

import (
	"context"
	"errors"
	"fmt"

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
		WithReplySubjecter(ReplySubjectInbox()),
	}
	opts = append(defaults, opts...)
	c := &clientImpl{}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type ClientOption = func(*clientImpl) *clientImpl

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

func NewTypedClient[A, R any](nc peanats.Connection, opts ...TypedClientOption[A, R]) TypedClient[A, R] {
	defaults := []TypedClientOption[A, R]{
		WithCodec[A, R](peanats.JsonCodec{}),
		WithTypedReplySubjecter[A, R](ReplySubjectInbox()),
	}
	opts = append(defaults, opts...)
	c := &typedClientImpl[A, R]{clientImpl: clientImpl{conn: nc}}
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

type TypedClientOption[A, R any] func(*typedClientImpl[A, R]) *typedClientImpl[A, R]

func WithCodec[A, R any](codec peanats.Codec) TypedClientOption[A, R] {
	return func(c *typedClientImpl[A, R]) *typedClientImpl[A, R] {
		c.codec = codec
		return c
	}
}

func WithTypedReplySubjecter[A, R any](rs ReplySubjecter) TypedClientOption[A, R] {
	return func(c *typedClientImpl[A, R]) *typedClientImpl[A, R] {
		c.rs = rs
		return c
	}
}

type typedClientImpl[A, R any] struct {
	clientImpl
	codec peanats.Codec
}

func (t *typedClientImpl[A, R]) WithCodec(codec peanats.Codec) TypedClient[A, R] {
	t.codec = codec
	return t
}

func (t *typedClientImpl[A, R]) Start(ctx context.Context, subj string, arg *A) (TypedReceiver[R], error) {
	data, err := t.codec.Encode(arg)
	if err != nil {
		return nil, err
	}
	rcv, err := t.clientImpl.Start(ctx, subj, data)
	if err != nil {
		return nil, err
	}
	return &typedReceiverImpl[R]{Receiver: rcv, codec: t.codec}, nil
}
