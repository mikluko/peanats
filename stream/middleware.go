package stream

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

// Middleware is a middleware that forms handler publication into a stream of messages.
//
// 1. As a first step, it sends an message to the original reply subject containing the subject of the stream.
// 2. All subsequent messages published from inner handler are sent to the stream subject.
// 3. When the handler returns, middleware sends an empty message to the stream subject marking the end of the stream.
func Middleware(next peanats.Handler) peanats.Handler {
	return &streamHandler{
		next: next,
	}
}

const (
	HeaderUID      = "Stream-UID"
	HeaderSequence = "Stream-Sequence"
	HeaderControl  = "Stream-Control"

	HeaderControlAck     = "ack"
	HeaderControlProceed = "proceed"
	HeaderControlDone    = "done"
)

type streamHandler struct {
	next peanats.Handler
}

func (h *streamHandler) validate(rq peanats.Request) error {
	if rq.Reply() == "" {
		return &peanats.Error{
			Code:    http.StatusBadRequest,
			Message: "reply subject is not set",
		}
	}
	if uid := rq.Header().Get(HeaderUID); uid == "" {
		return &peanats.Error{
			Code:    http.StatusBadRequest,
			Message: "UID header is not set",
		}
	}
	return nil
}

func (h *streamHandler) ack(pub peanats.PublisherMsg, rq peanats.Request) error {
	msg := nats.NewMsg(rq.Reply())
	msg.Header.Add(HeaderControl, HeaderControlAck)
	msg.Header.Add(HeaderUID, rq.Header().Get(HeaderUID))
	err := pub.PublishMsg(msg)
	if err != nil {
		return &peanats.Error{
			Code:    http.StatusInternalServerError,
			Message: "failed to publish acknowledgment",
			Cause:   err,
		}
	}
	return nil
}

func (h *streamHandler) done(pub peanats.PublisherMsg, rq peanats.Request) error {
	msg := nats.NewMsg(rq.Reply())
	msg.Header.Add(HeaderUID, rq.Header().Get(HeaderUID))
	msg.Header.Add(HeaderControl, HeaderControlDone)
	err := pub.PublishMsg(msg)
	if err != nil {
		return &peanats.Error{
			Code:    http.StatusInternalServerError,
			Message: "Failed to publish done",
			Cause:   err,
		}
	}
	return nil
}

func (h *streamHandler) Serve(pub peanats.Publisher, rq peanats.Request) error {
	err := h.validate(rq)
	if err != nil {
		return err
	}
	err = h.ack(pub, rq)
	if err != nil {
		return err
	}

	pubstream := &streamPublisherImpl{Publisher: pub.WithSubject(rq.Reply())}
	pubstream.Header().Add(HeaderUID, rq.Header().Get(HeaderUID))
	err = h.next.Serve(pubstream, rq)
	if err != nil {
		return err
	}
	if !pubstream.done {
		return h.done(pub, rq)
	}
	return nil
}

type streamPublisherImpl struct {
	peanats.Publisher
	seq  int
	done bool
}

func (p *streamPublisherImpl) WithSubject(subject string) peanats.Publisher {
	return &streamPublisherImpl{p.Publisher.WithSubject(subject), p.seq, p.done}
}

func (p *streamPublisherImpl) Publish(data []byte) error {
	if p.done {
		panic("protocol violation: publish after done")
	}
	defer func() { p.seq++ }()
	hdr := p.Header()
	switch v := hdr.Get(HeaderControl); v {
	case "":
		hdr.Set(HeaderControl, HeaderControlProceed)
		hdr.Set(HeaderSequence, strconv.Itoa(p.seq))
		defer hdr.Del(HeaderControl)
		defer hdr.Del(HeaderSequence)
	case HeaderControlProceed:
		hdr.Set(HeaderSequence, strconv.Itoa(p.seq))
		defer hdr.Del(HeaderSequence)
	case HeaderControlDone:
		if len(data) != 0 {
			panic("protocol violation: done message must not contain data")
		}
		p.done = true
	default:
		panic(fmt.Sprintf("protocol violation: invalid control header value: %s", v))
	}
	return p.Publisher.Publish(data)
}
