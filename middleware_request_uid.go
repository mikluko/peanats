package peanats

import "github.com/nats-io/nuid"

const HeaderRequestUID = "Request-UID"

type UIDGenerator interface {
	Next() string
}

func MakeRequestUIDMiddleware(gen UIDGenerator) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			if req.Header().Get(HeaderRequestUID) == "" {
				req.Header().Set(HeaderRequestUID, gen.Next())
			}
			return next.Serve(pub, req)
		})
	}
}

type globalNUIDGenerator struct{}

func (g *globalNUIDGenerator) Next() string {
	return nuid.Next()
}

var RequestUIDMiddleware = MakeRequestUIDMiddleware(&globalNUIDGenerator{})
