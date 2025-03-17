package peanats

import (
	"context"
	"net/textproto"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Header = textproto.MIMEHeader

type Msg interface {
	Subject() string
	Header() Header
	Data() []byte
}

type Metadatable interface {
	Metadata() (*jetstream.MsgMetadata, error)
}

type Respondable interface {
	Respond(context.Context, any) error
	RespondHeader(context.Context, any, Header) error
	RespondMsg(context.Context, Msg) error
}

type Ackable interface {
	Ack(context.Context) error
	Nak(context.Context) error
	Term(context.Context) error
	TermWithReason(context.Context, string) error
	InProgress(context.Context) error
}

type MsgJetstream interface {
	Msg
	Metadatable
	Ackable
}

type MsgHandler interface {
	HandleMsg(context.Context, Msg) error
}

type MsgHandlerFunc func(context.Context, Msg) error

func (f MsgHandlerFunc) HandleMsg(ctx context.Context, m Msg) error {
	return f(ctx, m)
}

type MsgMiddleware func(MsgHandler) MsgHandler

func ChainMsgMiddleware(h MsgHandler, mw ...MsgMiddleware) MsgHandler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

func NewMsg(m *nats.Msg) Msg {
	return &msgImpl{m}
}

type msgImpl struct {
	*nats.Msg
}

func (m *msgImpl) Subject() string {
	return m.Msg.Subject
}

func (m *msgImpl) Header() Header {
	return Header(m.Msg.Header)
}

func (m *msgImpl) Data() []byte {
	return m.Msg.Data
}

func (m *msgImpl) Respond(ctx context.Context, x any) error {
	return m.RespondHeader(ctx, x, nil)
}

func (m *msgImpl) RespondHeader(_ context.Context, x any, header Header) error {
	if header == nil {
		header = make(Header)
	}
	data, err := MarshalHeader(x, header)
	if err != nil {
		return err
	}
	return m.Msg.RespondMsg(&nats.Msg{
		Data:   data,
		Header: nats.Header(header),
	})
}

func (m *msgImpl) RespondMsg(_ context.Context, msg Msg) error {
	return m.Msg.RespondMsg(&nats.Msg{
		Data:   msg.Data(),
		Header: nats.Header(msg.Header()),
	})
}

func NewJetstream(m jetstream.Msg) MsgJetstream {
	return &msgJetstreamImpl{m}
}

type msgJetstreamImpl struct {
	jetstream.Msg
}

func (m *msgJetstreamImpl) Ack(_ context.Context) error {
	return m.Msg.Ack()
}

func (m *msgJetstreamImpl) Nak(_ context.Context) error {
	return m.Msg.Nak()
}

func (m *msgJetstreamImpl) Term(_ context.Context) error {
	return m.Msg.Term()
}

func (m *msgJetstreamImpl) TermWithReason(_ context.Context, s string) error {
	return m.Msg.TermWithReason(s)
}

func (m *msgJetstreamImpl) InProgress(_ context.Context) error {
	return m.Msg.InProgress()
}

func (m *msgJetstreamImpl) Header() Header {
	return Header(m.Msg.Headers())
}
