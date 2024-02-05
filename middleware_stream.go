package peanats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nuid"
	"net/http"
	"strconv"
)

// StreamMiddleware is a middleware that forms publications from the handler into a stream of messages.
//
// 1. As a first step, it sends an message to the original reply subject containing the subject of the stream.
// 2. All subsequent messages published from inner handler are sent to the stream subject.
// 3. When the handler returns, middleware sends an empty message to the stream subject marking the end of the stream.
func StreamMiddleware(next Handler) Handler {

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
	next Handler
}

func (h *streamHandler) ack(pub Publisher, rq Request) error {
	if rq.Reply() == "" {
		return &Error{
			Code:    http.StatusBadRequest,
			Message: "Reply subject is not set",
		}
	}
	msg := nats.NewMsg(rq.Reply())
	msg.Data = []byte(h.subj)
	msg.Header.Add(HeaderStreamUID, h.uid)
	err := pub.PublishMsg(msg)
	if err != nil {
		return &Error{
			Code:    http.StatusInternalServerError,
			Message: "Failed to publish ack",
			Cause:   err,
		}
	}
	return nil
}

func (h *streamHandler) done(pub Publisher) error {
	msg := nats.NewMsg(h.subj)
	msg.Header.Add(HeaderStreamUID, h.uid)
	msg.Header.Add(HeaderStreamControl, StreamControlDone)
	err := pub.PublishMsg(msg)
	if err != nil {
		return &Error{
			Code:    http.StatusInternalServerError,
			Message: "Failed to publish done",
			Cause:   err,
		}
	}
	return nil
}

func (h *streamHandler) Serve(pub Publisher, rq Request) error {
	err := h.ack(pub, rq)
	if err != nil {
		return err
	}
	pubstream := &streamPublisherImpl{Publisher: &publisherImpl{PublisherMsg: pub, subject: h.subj}}
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
	Publisher
	seq  int
	done bool
}

func (p *streamPublisherImpl) Publish(data []byte) error {
	p.seq++
	if p.Header().Get(HeaderStreamControl) != StreamControlDone {
		p.Header().Set(HeaderStreamControl, StreamControlProceed)
		defer p.Header().Del(HeaderStreamControl)
	} else {
		p.done = true
	}

	p.Header().Set(HeaderStreamSequence, strconv.Itoa(p.seq))
	defer p.Header().Del(HeaderStreamSequence)

	return p.Publisher.Publish(data)
}
