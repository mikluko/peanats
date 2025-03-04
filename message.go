package peanats

import (
	"context"
	"net/textproto"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Header = textproto.MIMEHeader

// Handler interface defines basic NATS message handler.
type Handler interface {
	Handle(context.Context, Dispatcher, Message)
}

type HandlerFunc func(context.Context, Dispatcher, Message)

func (f HandlerFunc) Handle(ctx context.Context, d Dispatcher, m Message) {
	f(ctx, d, m)
}

type Message interface {
	Subject() string
	Header() Header
	Data() []byte
}

func NewMessage(m *nats.Msg) Message {
	return &messageImpl{m}
}

type messageImpl struct {
	m *nats.Msg
}

func (m *messageImpl) Subject() string {
	return m.m.Subject
}

func (m *messageImpl) Header() Header {
	return Header(m.m.Header)
}

func (m *messageImpl) Data() []byte {
	return m.m.Data
}

type ResponseMessage interface {
	Header() Header
	Data() []byte
}

type RequestMessage interface {
	Message
	Reply() string
	Respond(context.Context, ResponseMessage) error
}

func NewRequestMessage(m *nats.Msg) RequestMessage {
	return &requestMessageImpl{messageImpl{m: m}}
}

type requestMessageImpl struct {
	messageImpl
}

func (r *requestMessageImpl) Reply() string {
	return r.m.Reply
}

func (r *requestMessageImpl) Respond(_ context.Context, m ResponseMessage) error {
	return r.m.RespondMsg(&nats.Msg{
		Header: nats.Header(m.Header()),
		Data:   m.Data(),
	})
}

type JetstreamMessage interface {
	Message
	Ack(context.Context, ...AckOption) error
	Nak(context.Context, ...NakOption) error
	Term(context.Context, ...TermOption) error
	InProgress(context.Context) error
}

type AckOption func(*ackParams)

type ackParams struct{}

type NakOption func(*nakParams)

type nakParams struct{}

type TermOption func(*termParams)

type termParams struct{}

func NewJetstreamMessage(m jetstream.Msg) JetstreamMessage {
	return &jetstreamMessageImpl{m}
}

type jetstreamMessageImpl struct {
	m jetstream.Msg
}

func (j *jetstreamMessageImpl) Subject() string {
	return j.m.Subject()
}

func (j *jetstreamMessageImpl) Header() Header {
	return Header(j.m.Headers())
}

func (j *jetstreamMessageImpl) Data() []byte {
	return j.m.Data()
}

func (j *jetstreamMessageImpl) Ack(_ context.Context, _ ...AckOption) error {
	return j.m.Ack()
}

func (j *jetstreamMessageImpl) Nak(_ context.Context, _ ...NakOption) error {
	return j.m.Nak()
}

func (j *jetstreamMessageImpl) Term(_ context.Context, _ ...TermOption) error {
	return j.m.Term()
}

func (j *jetstreamMessageImpl) InProgress(_ context.Context) error {
	return j.m.InProgress()
}
