package peanats

import (
	"context"
	"errors"
	"fmt"
	"net/textproto"
	"regexp"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Message interface defines a basic message.
type Message interface {
	Subject() string
	Header() textproto.MIMEHeader
	Data() []byte
	Respond(context.Context, any, ...RespondOption) error
}

type Metadatable interface {
	Metadata() jetstream.MsgMetadata
}

type Ackable interface {
	Ack(context.Context) error
	Nak(context.Context) error
	Term(context.Context, ...TermOption) error
	InProgress(context.Context) error
}

type MessageJetstream interface {
	Message
	Metadatable
	Ackable
}

type RespondOption func(params *respondParams)

type respondParams struct {
	pub    Publisher
	header textproto.MIMEHeader
	ctype  ContentType
}

func RespondPublisher(pub Publisher) RespondOption {
	return func(p *respondParams) {
		p.pub = pub
	}
}

func RespondHeader(header textproto.MIMEHeader) RespondOption {
	return func(p *respondParams) {
		p.header = header
	}
}

func RespondContentType(ctype ContentType) RespondOption {
	return func(p *respondParams) {
		p.ctype = ctype
	}
}

type messageImpl struct {
	*nats.Msg
}

func (m *messageImpl) Subject() string {
	return m.Msg.Subject
}

func (m *messageImpl) Header() textproto.MIMEHeader {
	return textproto.MIMEHeader(m.Msg.Header)
}

func (m *messageImpl) Data() []byte {
	return m.Msg.Data
}

func (m *messageImpl) Respond(ctx context.Context, x any, opts ...RespondOption) error {
	p := respondParams{
		header: make(textproto.MIMEHeader),
		ctype:  ContentTypeHeader(m.Header()),
	}
	for _, opt := range opts {
		opt(&p)
	}
	if p.pub == nil {
		codec, err := CodecContentType(p.ctype)
		if err != nil {
			return err
		}
		codec.SetHeader(p.header)
		data, err := codec.Encode(x)
		if err != nil {
			return err
		}
		return m.Msg.RespondMsg(&nats.Msg{
			Header: nats.Header(p.header),
			Data:   data,
		})
	}
	return p.pub.Publish(ctx, m.Reply, x, PublishHeader(p.header), PublishContentType(p.ctype))
}

type messageJetstreamImpl struct {
	jetstream.Msg
}

func (m *messageJetstreamImpl) Subject() string {
	return m.Msg.Subject()
}

func (m *messageJetstreamImpl) Header() textproto.MIMEHeader {
	return textproto.MIMEHeader(m.Msg.Headers())
}

func (m *messageJetstreamImpl) Metadata() jetstream.MsgMetadata {
	meta, _ := m.Msg.Metadata()
	return *meta
}

func (m *messageJetstreamImpl) Respond(ctx context.Context, x any, opts ...RespondOption) error {
	p := respondParams{
		header: make(textproto.MIMEHeader),
		ctype:  ContentTypeHeader(m.Header()),
	}
	for _, opt := range opts {
		opt(&p)
	}
	if p.pub == nil {
		return errors.New("publisher is not set")
	}
	return p.pub.Publish(ctx, m.Reply(), x, PublishHeader(p.header), PublishContentType(p.ctype))
}

func (m *messageJetstreamImpl) Ack(_ context.Context) error {
	return m.Msg.Ack()
}

func (m *messageJetstreamImpl) Nak(_ context.Context) error {
	return m.Msg.Nak()
}

var _ MessageJetstream = (*messageJetstreamImpl)(nil)

type TermOption func(*termParams)

func TermReason(reason string) TermOption {
	return func(p *termParams) {
		p.reason = reason
	}
}

type termParams struct {
	reason string
}

func (m *messageJetstreamImpl) Term(_ context.Context, opts ...TermOption) error {
	p := termParams{}
	for _, opt := range opts {
		opt(&p)
	}
	if p.reason == "" {
		return m.Msg.Term()
	} else {
		return m.Msg.TermWithReason(p.reason)
	}
}

func (m *messageJetstreamImpl) InProgress(_ context.Context) error {
	return m.Msg.InProgress()
}

var _ Ackable = (*messageJetstreamImpl)(nil)

// MessageHandler interface defines basic NATS message handler.
type MessageHandler interface {
	Handle(context.Context, Message) error
}

type MessageHandlerFunc func(context.Context, Message) error

func (f MessageHandlerFunc) Handle(ctx context.Context, m Message) error {
	return f(ctx, m)
}

type MessageMiddleware func(MessageHandler) MessageHandler

func ChainMessageMiddleware(h MessageHandler, mw ...MessageMiddleware) MessageHandler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

type Route interface {
	Match(string) bool
	MessageHandler
}

func NewRoute(handle MessageHandler, patterns ...string) Route {
	exps := make([]regexp.Regexp, 0, len(patterns))
	for _, pat := range patterns {
		pat = regexp.QuoteMeta(pat)
		if strings.Contains(pat, "*") {
			pat = strings.ReplaceAll(pat, "\\*", "[^.]+")
		}
		if strings.HasSuffix(pat, ">") {
			pat = strings.TrimSuffix(pat, ">") + "(.+)"
		}
		exps = append(exps, *regexp.MustCompile("^" + pat + "$"))
	}
	return &routeImpl{handle, exps}
}

type routeImpl struct {
	MessageHandler
	exps []regexp.Regexp
}

func (r *routeImpl) Match(subj string) bool {
	for i := range r.exps {
		if r.exps[i].MatchString(subj) {
			return true
		}
	}
	return false
}

// Muxer is a handler multiplexer
type Muxer interface {
	MessageHandler
	Add(Route)
}

func NewMuxer(routes ...Route) Muxer {
	return &muxerImpl{
		routes: routes,
	}
}

var ErrNotFound = fmt.Errorf("no route found")

type muxerImpl struct {
	routes []Route
}

func (r *muxerImpl) Handle(ctx context.Context, m Message) error {
	for _, route := range r.routes {
		if route.Match(m.Subject()) {
			return route.Handle(ctx, m)
		}
	}
	return fmt.Errorf("%w: %s", ErrNotFound, m.Subject())
}

func (r *muxerImpl) Add(route Route) {
	r.routes = append(r.routes, route)
}
