package peanats

import (
	"github.com/nats-io/nuid"
)

const HeaderRequestUID = "Request-UID"

type UIDGenerator interface {
	Next() string
}

func RequestUIDMiddleware(next Handler) Handler {
	return HandlerFunc(func(pub Publisher, req Request) error {
		header := req.Header()
		uid := header.Get(HeaderRequestUID)
		if uid == "" {
			uid = nuid.Next()
			header.Set(HeaderRequestUID, uid)
		}
		return next.Serve(&requestUIDPublisher{pub, uid}, req)
	})
}

type requestUIDPublisher struct {
	Publisher
	uid string
}

func (p *requestUIDPublisher) WithSubject(subject string) Publisher {
	return &requestUIDPublisher{p.Publisher.WithSubject(subject), p.uid}
}

func (p *requestUIDPublisher) Publish(data []byte) error {
	if p.Header().Get(HeaderRequestUID) == "" {
		p.Header().Set(HeaderRequestUID, p.uid)
	}
	return p.Publisher.Publish(data)
}
