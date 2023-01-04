package peanats

import "github.com/nats-io/nats.go"

func MakePublishSubjectMiddleware(conn *nats.Conn, subject string) Middleware {
	return func(h Handler) Handler {
		return HandlerFunc(func(pub Publisher, req Request) error {
			return h.Serve(&subjectPublisher{pub, conn, subject}, req)
		})
	}
}
