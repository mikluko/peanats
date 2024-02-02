package peanats

import (
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nuid"
)

// MakeAccessLogMiddleware returns a middleware that logs the request and response of a handler.
//
// WARNING: the logger is inefficient and should not be used in production.
// TODO: replace with efficient implementation.
func MakeAccessLogMiddleware(opts ...AccessLogMiddlewareOption) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			uid := req.Header().Get(HeaderRequestUID)
			if uid == "" {
				uid = nuid.Next()
			}
			p := loggingPublisher{
				Publisher: pub,
				started:   time.Now(),
				subj:      req.Subject(),
				uid:       uid,
			}
			for i := range opts {
				opts[i](&p)
			}
			if p.log == nil {
				AccessLogMiddlewareWithWriter(log.Writer())(&p)
			}
			err := next.Serve(&p, req)
			if !p.hasPublished {
				p.log.Println(p.uid, p.subj, time.Since(p.started).Milliseconds(), "noop")
			}
			return err
		})
	}
}

type AccessLogger interface {
	Println(v ...any)
}

type AccessLogMiddlewareOption func(*loggingPublisher)

func AccessLogMiddlewareWithLogger(log AccessLogger) AccessLogMiddlewareOption {
	return func(p *loggingPublisher) {
		p.log = log
	}
}

func AccessLogMiddlewareWithWriter(w io.Writer) AccessLogMiddlewareOption {
	return AccessLogMiddlewareWithLogger(log.New(w, "", log.Ldate|log.Ltime|log.LUTC))
}

type loggingPublisher struct {
	Publisher
	log          AccessLogger
	uid          string
	subj         string
	started      time.Time
	hasPublished bool
}

func (p *loggingPublisher) Publish(data []byte) error {
	p.hasPublished = true
	err := p.Publisher.Publish(data)
	if err != nil {
		return err
	}
	bits := []string{p.uid, p.subj,
		strconv.Itoa(int(time.Since(p.started).Milliseconds())),
		"pub",
		p.Subject(),
		strconv.Itoa(len(data)),
		p.errCode(),
		p.errMsg(),
	}
	p.log.Println(strings.TrimSpace(strings.Join(bits, " ")))
	return nil
}

func (p *loggingPublisher) errCode() string {
	return p.Header().Get(HeaderErrorCode)
}

func (p *loggingPublisher) errMsg() string {
	return p.Header().Get(HeaderErrorMessage)
}
