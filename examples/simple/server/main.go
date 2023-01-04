package main

import (
	"github.com/mikluko/peanats"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	srv := peanats.Server{
		ListenSubjects: []string{"peanuts.simple.requests"},
		Conn:           nc,
		Handler: peanats.ChainMiddleware(
			peanats.HandlerFunc(handle),
			peanats.MakePublishSubjectMiddleware(nc, "peanuts.simple.results"),
		),
	}

	err = srv.Start()
	if err != nil {
		panic(err)
	}
	err = srv.Wait()
	if err != nil {
		panic(err)
	}
}

func handle(pub peanats.Publisher, req peanats.Request) error {
	if ack, ok := pub.(peanats.Acker); ok {
		err := ack.Ack(nil)
		if err != nil {
			return err
		}
	}
	return pub.Publish(req.Data())
}
