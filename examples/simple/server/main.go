package main

import (
	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	srv := peanats.Server{
		ListenSubjects: []string{"peanuts.simple.requests"},
		Conn:           peanats.NATS(nc),
		Handler: peanats.ChainMiddleware(
			peanats.HandlerFunc(handle),
			peanats.MakeAckMiddleware(nc, peanats.AckMiddlewareWithPayload([]byte("ACK"))),
			peanats.MakePublishSubjectMiddleware(nc, "peanuts.simple.results"),
		),
	}

	err = srv.Start()
	if err != nil {
		panic(err)
	}
	srv.Wait()
}

func handle(pub peanats.Publisher, req peanats.Request) error {
	return pub.Publish(req.Data())
}
