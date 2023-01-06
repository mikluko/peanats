package peanats

import (
	"github.com/nats-io/nuid"
)

const HeaderRequestUID = "Request-UID"

type UIDGenerator interface {
	Next() string
}

func MakeRequestUIDMiddleware(gen UIDGenerator) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			header := req.Header()
			uid := header.Get(HeaderRequestUID)
			if uid == "" {
				uid = gen.Next()
				header.Set(HeaderRequestUID, uid)
			}
			return next.Serve(&requestUIDPublisher{pub, uid}, req)
		})
	}
}

type globalNUIDGenerator struct{}

func (g *globalNUIDGenerator) Next() string {
	return nuid.Next()
}

var RequestUIDMiddleware = MakeRequestUIDMiddleware(&globalNUIDGenerator{})

type requestUIDPublisher struct {
	Publisher
	uid string
}

func (p *requestUIDPublisher) Ack(data []byte) error {
	ack, ok := p.Publisher.(AckPublisher)
	if !ok {
		return nil
	}
	if ack.Header().Get(HeaderRequestUID) == "" {
		ack.Header().Set(HeaderRequestUID, p.uid)
	}
	return ack.Ack(data)
}

func (p *requestUIDPublisher) Publish(data []byte) error {
	if p.Header().Get(HeaderRequestUID) == "" {
		p.Header().Set(HeaderRequestUID, p.uid)
	}
	return p.Publisher.Publish(data)
}
