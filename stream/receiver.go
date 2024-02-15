package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

type Receiver interface {
	UID() string
	Sequence() int
	Receive(context.Context) (*nats.Msg, error)
	ReceiveAll(context.Context) ([]*nats.Msg, error)
}

type receiverImpl struct {
	sub peanats.Subscription
	uid string
	seq int
}

func (r *receiverImpl) UID() string {
	return r.uid
}

func (r *receiverImpl) Sequence() int {
	return r.seq
}

func (r *receiverImpl) Receive(ctx context.Context) (*nats.Msg, error) {
	msg, err := r.sub.NextMsg(ctx)
	if err != nil {
		return nil, err
	}
	if uid := msg.Header.Get(HeaderUID); uid != r.uid {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "unexpected UID")
	}
	if msg.Header.Get(HeaderControl) == HeaderControlDone {
		if len(msg.Data) != 0 {
			return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "unexpected data in done message")
		}
		return nil, io.EOF
	}
	if seq := msg.Header.Get(HeaderSequence); seq == "" {
		return nil, fmt.Errorf("%w: %s: %w", ErrProtocolViolation, "sequence header is empty or absent", err)
	} else if seq != strconv.Itoa(r.seq) {
		return nil, fmt.Errorf("%w: %s", ErrProtocolViolation, "sequence number expectation violated")
	}
	r.seq++
	return msg, err
}

func (r *receiverImpl) ReceiveAll(ctx context.Context) ([]*nats.Msg, error) {
	var msgs []*nats.Msg
	for {
		msg, err := r.Receive(ctx)
		if errors.Is(err, io.EOF) {
			return msgs, nil
		} else if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
}

type TypedReceiver[T any] interface {
	UID() string
	Sequence() int
	Receive(context.Context) (*T, error)
	ReceiveAll(context.Context) ([]*T, error)
}

type typedReceiverImpl[T any] struct {
	Receiver
	codec peanats.Codec
}

func (r *typedReceiverImpl[T]) decode(msg *nats.Msg) (*T, error) {
	obj := new(T)
	if err := r.codec.Decode(msg.Data, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *typedReceiverImpl[T]) Receive(ctx context.Context) (*T, error) {
	data, err := r.Receiver.Receive(ctx)
	if err != nil {
		return nil, err
	}
	return r.decode(data)
}

func (r *typedReceiverImpl[T]) ReceiveAll(ctx context.Context) ([]*T, error) {
	var objs []*T
	for {
		obj, err := r.Receive(ctx)
		if errors.Is(err, io.EOF) {
			return objs, nil
		} else if err != nil {
			return nil, err
		}
		objs = append(objs, obj)
	}
}
