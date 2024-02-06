package stream

import (
	"github.com/mikluko/peanats"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
	"net/http"
	"strconv"
)

// Middleware is a middleware that forms handler publication into a stream of messages.
//
// 1. As a first step, it sends an message to the original reply subject containing the subject of the stream.
// 2. All subsequent messages published from inner handler are sent to the stream subject.
// 3. When the handler returns, middleware sends an empty message to the stream subject marking the end of the stream.
func Middleware(next peanats.Handler) peanats.Handler {

	uid := nuid.Next()

	var b [streamPrefixLen + nuidSize]byte
	copy(b[:streamPrefixLen], StreamPrefix)
	copy(b[streamPrefixLen:], uid)

	return &streamHandler{
		uid:  uid,
		subj: string(b[:]),
		next: next,
	}
}

const (
	HeaderStreamUID      = "Stream-UID"
	HeaderStreamSequence = "Stream-Sequence"

	HeaderStreamControl  = "Stream-Control"
	StreamControlProceed = "proceed"
	StreamControlDone    = "done"

	StreamPrefix    = "_STREAM."
	streamPrefixLen = len(StreamPrefix)
	nuidSize        = 22
)

type streamHandler struct {
	uid  string
	subj string
	next peanats.Handler
}

func (h *streamHandler) ack(pub peanats.Publisher, rq peanats.Request) error {
	if rq.Reply() == "" {
		return &peanats.Error{
			Code:    http.StatusBadRequest,
			Message: "Reply subject is not set",
		}
	}
	msg := nats.NewMsg(rq.Reply())
	msg.Data = []byte(h.subj)
	msg.Header.Add(HeaderStreamUID, h.uid)
	err := pub.PublishMsg(msg)
	if err != nil {
		return &peanats.Error{
			Code:    http.StatusInternalServerError,
			Message: "Failed to publish ack",
			Cause:   err,
		}
	}
	return nil
}

func (h *streamHandler) done(pub peanats.Publisher) error {
	msg := nats.NewMsg(h.subj)
	msg.Header.Add(HeaderStreamUID, h.uid)
	msg.Header.Add(HeaderStreamControl, StreamControlDone)
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
	err := h.ack(pub, rq)
	if err != nil {
		return err
	}
	pubstream := &streamPublisherImpl{pub.WithSubject(h.subj), 0, false}
	pubstream.Header().Add(HeaderStreamUID, h.uid)
	err = h.next.Serve(pubstream, rq)
	if err != nil {
		return err
	}
	if !pubstream.done {
		return h.done(pubstream)
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
	p.seq++
	hdr := p.Header()
	if hdr.Get(HeaderStreamControl) != StreamControlDone {
		hdr.Set(HeaderStreamControl, StreamControlProceed)
		defer hdr.Del(HeaderStreamControl)
	} else {
		p.done = true
	}

	hdr.Set(HeaderStreamSequence, strconv.Itoa(p.seq))
	defer hdr.Del(HeaderStreamSequence)

	return p.Publisher.Publish(data)
}
