package peanats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mikluko/peanats/internal/xargpool"
)

type Arg[T any] interface {
	Msg
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

var ErrArgumentUnmarshalFailed = errors.New("failed to unmarshal message into argument")

func MsgHandlerFromArgHandler[T any](h ArgHandler[T]) MsgHandler {
	pool := xargpool.New[T]()
	return MsgHandlerFunc(func(ctx context.Context, m Msg) error {
		y := pool.Acquire(ctx)
		x := y.Value()
		defer y.Release()

		err := UnmarshalHeader(m.Data(), x, m.Header())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrArgumentUnmarshalFailed, err)
		}

		return h.HandleArg(ctx, NewArg(m, x))
	})
}

var _ Arg[any] = (*argImpl[any])(nil)

func NewArg[T any](msg Msg, x *T) Arg[T] {
	switch msg.(type) {
	case Ackable:
		return &argAckableImpl[T]{msg, x}
	case Respondable:
		return &argRespondableImpl[T]{msg, x}
	default:
		return &argImpl[T]{msg, x}
	}
}

type argImpl[T any] struct {
	Msg
	x *T
}

func (a argImpl[T]) Value() *T {
	return a.x
}

type argRespondableImpl[T any] struct {
	Msg
	x *T
}

func (a *argRespondableImpl[T]) Respond(ctx context.Context, x any) error {
	return a.Msg.(Respondable).Respond(ctx, x)
}

func (a *argRespondableImpl[T]) RespondHeader(ctx context.Context, x any, header Header) error {
	return a.Msg.(Respondable).RespondHeader(ctx, x, header)
}

func (a *argRespondableImpl[T]) RespondMsg(ctx context.Context, msg Msg) error {
	return a.Msg.(Respondable).RespondMsg(ctx, msg)
}

func (a *argRespondableImpl[T]) Value() *T {
	return a.x
}

var _ Respondable = (*argRespondableImpl[any])(nil)

type argAckableImpl[T any] struct {
	Msg
	x *T
}

func (a *argAckableImpl[T]) Ack(ctx context.Context) error {
	return a.Msg.(Ackable).Ack(ctx)
}

func (a *argAckableImpl[T]) Nak(ctx context.Context) error {
	return a.Msg.(Ackable).Nak(ctx)
}

func (a *argAckableImpl[T]) NackWithDelay(ctx context.Context, d time.Duration) error {
	return a.Msg.(Ackable).NackWithDelay(ctx, d)
}

func (a *argAckableImpl[T]) Term(ctx context.Context) error {
	return a.Msg.(Ackable).Term(ctx)
}

func (a *argAckableImpl[T]) TermWithReason(ctx context.Context, reason string) error {
	return a.Msg.(Ackable).TermWithReason(ctx, reason)
}

func (a *argAckableImpl[T]) InProgress(ctx context.Context) error {
	return a.Msg.(Ackable).InProgress(ctx)
}

func (a *argAckableImpl[T]) Value() *T {
	return a.x
}

var _ Ackable = (*argAckableImpl[any])(nil)
