package peanats

import "github.com/nats-io/nats.go"

func MakePublishRespondMiddleware(conn *nats.Conn) Middleware {
	return func(h Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			return h.Serve(&subjectPublisher{pub, conn, req.Reply()}, req)
		})
	}
}
