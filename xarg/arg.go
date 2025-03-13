package xarg

import (
	"context"
	"errors"
	"fmt"

	"github.com/mikluko/peanats/internal/xargpool"
	"github.com/mikluko/peanats/xenc"
	"github.com/mikluko/peanats/xmsg"
)

type Arg[T any] interface {
	xmsg.Msg
	Value() *T
}

// ArgHandler interface defines a typed handler.
type ArgHandler[T any] interface {
	HandleArg(context.Context, Arg[T]) error
}

// ArgHandlerFunc is an adapter to allow the use of ordinary functions as Handler.
type ArgHandlerFunc[T any] func(context.Context, Arg[T]) error

func (f ArgHandlerFunc[T]) HandleArg(ctx context.Context, a Arg[T]) error {
	return f(ctx, a)
}

var ErrArgumentDecodeFailed = errors.New("failed to decode message into argument")

func ArgMsgHandler[T any](h ArgHandler[T]) xmsg.MsgHandler {
	pool := xargpool.New[T]()
	return xmsg.MsgHandlerFunc(func(ctx context.Context, m xmsg.Msg) error {
		codec, err := xenc.CodecContentType(xenc.ContentTypeHeader(m.Header()))
		if err != nil {
			return err
		}

		y := pool.Acquire(ctx)
		x := y.Value()
		defer y.Release()

		err = codec.Decode(m.Data(), x)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrArgumentDecodeFailed, err)
		}

		return h.HandleArg(ctx, New(m, x))
	})
}

var _ Arg[any] = (*argImpl[any])(nil)

func New[T any](msg xmsg.Msg, x *T) Arg[T] {
	switch msg.(type) {
	case xmsg.Ackable:
		return &argAckableImpl[T]{msg, x}
	case xmsg.Respondable:
		return &argRespondableImpl[T]{msg, x}
	default:
		return &argImpl[T]{msg, x}
	}
}

type argImpl[T any] struct {
	xmsg.Msg
	x *T
}

func (a argImpl[T]) Value() *T {
	return a.x
}

type argRespondableImpl[T any] struct {
	xmsg.Msg
	x *T
}

func (a *argRespondableImpl[T]) Respond(ctx context.Context, x any) error {
	return a.Msg.(xmsg.Respondable).Respond(ctx, x)
}

func (a *argRespondableImpl[T]) RespondHeader(ctx context.Context, x any, header xmsg.Header) error {
	return a.Msg.(xmsg.Respondable).RespondHeader(ctx, x, header)
}

func (a *argRespondableImpl[T]) RespondMsg(ctx context.Context, msg xmsg.Msg) error {
	return a.Msg.(xmsg.Respondable).RespondMsg(ctx, msg)
}

func (a *argRespondableImpl[T]) Value() *T {
	return a.x
}

var _ xmsg.Respondable = (*argRespondableImpl[any])(nil)

type argAckableImpl[T any] struct {
	xmsg.Msg
	x *T
}

func (a *argAckableImpl[T]) Ack(ctx context.Context) error {
	return a.Msg.(xmsg.Ackable).Ack(ctx)
}

func (a *argAckableImpl[T]) Nak(ctx context.Context) error {
	return a.Msg.(xmsg.Ackable).Nak(ctx)
}

func (a *argAckableImpl[T]) Term(ctx context.Context) error {
	return a.Msg.(xmsg.Ackable).Term(ctx)
}

func (a *argAckableImpl[T]) TermWithReason(ctx context.Context, reason string) error {
	return a.Msg.(xmsg.Ackable).TermWithReason(ctx, reason)
}

func (a *argAckableImpl[T]) InProgress(ctx context.Context) error {
	return a.Msg.(xmsg.Ackable).InProgress(ctx)
}

func (a *argAckableImpl[T]) Value() *T {
	return a.x
}

var _ xmsg.Ackable = (*argAckableImpl[any])(nil)
