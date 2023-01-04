package peanats

import (
	"github.com/nats-io/nuid"
	"io"
	"log"
	"time"
)

func MakeAccessLogMiddleware(w io.Writer) Middleware {
	logger := log.New(w, "", log.Ldate|log.Ltime|log.LUTC)
	return func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			uid := req.Header().Get(HeaderRequestUID)
			if uid == "" {
				uid = nuid.Next()
			}
			p := loggingPublisher{
				Publisher: pub,
				log:       logger.Println,
				uid:       uid,
				subj:      req.Subject(),
				started:   time.Now(),
			}
			err := next.Serve(&p, req)
			if !p.hasPublished && !p.hasAcked {
				logger.Println(p.uid, p.subj, time.Since(p.started).Milliseconds(), "noop")
			}
			return err
		})
	}
}

type loggingPublisher struct {
	Publisher
	log          func(v ...any)
	uid          string
	subj         string
	started      time.Time
	hasAcked     bool
	hasPublished bool
}

func (p *loggingPublisher) Ack(data []byte) error {
	ack, ok := p.Publisher.(Acker)
	if !ok {
		return nil
	}
	err := ack.Ack(data)
	if err != nil {
		return err
	}
	p.hasAcked = true
	p.log(p.uid, p.subj, time.Since(p.started).Milliseconds(), "ack", len(data), p.errCode(), p.errMsg())
	return nil
}

func (p *loggingPublisher) Publish(data []byte) error {
	err := p.Publisher.Publish(data)
	if err != nil {
		return err
	}
	p.hasPublished = true
	p.log(p.uid, p.subj, time.Since(p.started).Milliseconds(), "pub", len(data), p.errCode(), p.errMsg())
	return nil
}

func (p *loggingPublisher) errCode() string {
	return p.Header().Get(HeaderErrorCode)
}

func (p *loggingPublisher) errMsg() string {
	return p.Header().Get(HeaderErrorMessage)
}
