package requester

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
	"github.com/mikluko/peanats/transport"
)

type RequestOption func(*requestParams)

type requestParams struct {
	header peanats.Header
}

func makeRequestParams(opts ...RequestOption) requestParams {
	p := requestParams{
		header: peanats.Header{codec.HeaderContentType: []string{codec.JSON.String()}},
	}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

// RequestHeader merges the provided headers with existing headers for the message.
// If the same header key exists, the new values are appended to the existing ones.
func RequestHeader(header peanats.Header) RequestOption {
	return func(p *requestParams) {
		for key, values := range header {
			for _, value := range values {
				p.header.Add(key, value)
			}
		}
	}
}

func RequestContentType(c codec.ContentType) RequestOption {
	return func(p *requestParams) {
		p.header.Set(codec.HeaderContentType, c.String())
	}
}

type Requester[RQ, RS any] interface {
	Request(context.Context, string, *RQ, ...RequestOption) (Response[RS], error)
	ResponseReceiver(context.Context, string, *RQ, ...ResponseReceiverOption) (ResponseReceiver[RS], error)
}

func New[RQ, RS any](nc transport.Conn) Requester[RQ, RS] {
	return &clientImpl[RQ, RS]{nc}
}

type clientImpl[RQ, RS any] struct {
	nc transport.Conn
}

func (c *clientImpl[RQ, RS]) Request(ctx context.Context, subj string, rq *RQ, opts ...RequestOption) (Response[RS], error) {
	p := makeRequestParams(opts...)
	data, err := codec.MarshalHeader(rq, p.header)
	if err != nil {
		return nil, err
	}
	msg, err := c.nc.Request(ctx, requestMessageImpl{subj: subj, header: p.header, data: data})
	if err != nil {
		return nil, err
	}
	rs := new(RS)
	err = codec.UnmarshalHeader(msg.Data(), rs, msg.Header())
	if err != nil {
		return nil, err
	}
	return &responseImpl[RS]{header: msg.Header(), payload: rs}, nil
}

func (c *clientImpl[RQ, RS]) ResponseReceiver(ctx context.Context, subj string, rq *RQ, opts ...ResponseReceiverOption) (ResponseReceiver[RS], error) {
	rcvParams := responseReceiverParams{
		buffer:    DefaultBuffer,
		skipper:   DefaultSkipper,
		proceeder: DefaultProceeder,
	}
	for _, opt := range opts {
		opt(&rcvParams)
	}
	reqParams := makeRequestParams(rcvParams.rqOpts...)
	data, err := codec.MarshalHeader(rq, reqParams.header)
	if err != nil {
		return nil, err
	}
	msg := requestMessageImpl{subj: subj, repl: nats.NewInbox(), header: reqParams.header, data: data}

	// message is ready, prepare response sequence subscription
	buf := make(chan peanats.Msg, rcvParams.buffer)
	sub, err := c.nc.SubscribeChan(ctx, msg.Reply(), buf)
	if err != nil {
		return nil, err
	}

	// send the request
	err = c.nc.Publish(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &responseReceiverImpl[RS]{buf: buf, sub: sub, skp: rcvParams.skipper, pdr: rcvParams.proceeder}, nil
}

type requestMessageImpl struct {
	subj   string
	repl   string
	header peanats.Header
	data   []byte
}

func (r requestMessageImpl) Subject() string {
	return r.subj
}

func (r requestMessageImpl) Reply() string {
	return r.repl
}

func (r requestMessageImpl) Header() peanats.Header {
	return r.header
}

func (r requestMessageImpl) Data() []byte {
	return r.data
}

type Response[T any] interface {
	Header() peanats.Header
	Value() *T
}

type responseImpl[T any] struct {
	header  peanats.Header
	payload *T
}

func (r *responseImpl[T]) Header() peanats.Header {
	return r.header
}

func (r *responseImpl[T]) Value() *T {
	return r.payload
}

type ResponseReceiver[T any] interface {
	Next(context.Context) (Response[T], error)
	Stop() error
}

type ResponseReceiverOption func(*responseReceiverParams)

type responseReceiverParams struct {
	buffer    uint
	skipper   Skipper
	proceeder Proceeder
	rqOpts    []RequestOption
}

const DefaultBuffer = 0

// ResponseReceiverBuffer sets the buffer size for the response receiver.
func ResponseReceiverBuffer(size uint) ResponseReceiverOption {
	return func(r *responseReceiverParams) {
		r.buffer = size
	}
}

// Skipper makes decision whether or not skip the message handler without
// interrupting the response sequence.
type Skipper interface {
	Skip(context.Context, peanats.Msg) (bool, error)
}

type skipperImpl struct{}

func (s *skipperImpl) Skip(_ context.Context, msg peanats.Msg) (bool, error) {
	return len(msg.Data()) == 0, nil
}

// Proceeder makes decision whether or not proceed with the response sequence.
// Decision is made late, after the message handler is invoked or skipped.
type Proceeder interface {
	Proceed(context.Context, peanats.Msg) (bool, error)
}

type proceederImpl struct{}

func (p *proceederImpl) Proceed(_ context.Context, msg peanats.Msg) (bool, error) {
	return len(msg.Data()) != 0, nil
}

var (
	// DefaultSkipper skips empty messages.
	DefaultSkipper Skipper = &skipperImpl{}

	// DefaultProceeder proceeds only on non-empty messages.
	DefaultProceeder Proceeder = &proceederImpl{}

	// The combination of default Skipper and Proceeder effectively create logic
	// where the first empty message is not processed by the handler and
	// terminates the response sequence.
	_ = 0
)

// ResponseReceiverProceeder sets the proceeder for the response sequence.
func ResponseReceiverProceeder(p Proceeder) ResponseReceiverOption {
	return func(r *responseReceiverParams) {
		r.proceeder = p
	}
}

// ResponseReceiverSkipper sets the skipper for the response sequence.
func ResponseReceiverSkipper(s Skipper) ResponseReceiverOption {
	return func(r *responseReceiverParams) {
		r.skipper = s
	}
}

// ResponseReceiverRequestOptions appends the set of request options for the
// request produced by the response receiver.
func ResponseReceiverRequestOptions(opts ...RequestOption) ResponseReceiverOption {
	return func(r *responseReceiverParams) {
		r.rqOpts = append(r.rqOpts, opts...)
	}
}

type responseReceiverImpl[T any] struct {
	msg     peanats.Msg
	buf     chan peanats.Msg
	sub     peanats.Unsubscriber
	skp     Skipper
	pdr     Proceeder
	proceed bool
	once    sync.Once
}

var (
	ErrSkip = errors.New("message skipped by the skipper")
	ErrOver = fmt.Errorf("%w: sequence is over", io.EOF)
)

func (r *responseReceiverImpl[T]) Next(ctx context.Context) (_ Response[T], err error) {
	r.once.Do(func() {
		r.proceed = true
	})
	if !r.proceed {
		return nil, ErrOver
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r.msg = <-r.buf:
			r.proceed, err = r.pdr.Proceed(ctx, r.msg)
			if err != nil {
				return nil, err
			}
			if skip, err := r.skp.Skip(ctx, r.msg); err != nil {
				return nil, err
			} else if !skip {
				x := new(T)
				err := codec.UnmarshalHeader(r.msg.Data(), x, r.msg.Header())
				if err != nil {
					return nil, err
				}
				return &responseImpl[T]{header: r.msg.Header(), payload: x}, nil
			} else { // skip == true
				if r.proceed {
					return nil, ErrSkip
				} else {
					return nil, ErrOver
				}
			}
		}
	}
}

func (r *responseReceiverImpl[T]) Stop() error {
	close(r.buf)
	return r.sub.Unsubscribe()
}
