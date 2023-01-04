package main

import (
	"bytes"
	"github.com/mikluko/peanats"
	"github.com/nats-io/nats.go"
	"math/rand"
	"time"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	mux := new(peanats.ServeMux)
	_ = mux.Handle(
		peanats.ChainMiddleware(handleFunc("handle A:"), peanats.MakePublishSubjectMiddleware(nc, "peanuts.mux.results.a")),
		"peanuts.mux.requests.a",
	)
	_ = mux.Handle(
		peanats.ChainMiddleware(handleFunc("handle B:"), peanats.MakePublishSubjectMiddleware(nc, "peanuts.mux.results.b")),
		"peanuts.mux.requests.b",
	)

	srv := peanats.Server{
		ListenSubjects: []string{"peanuts.mux.requests.>"},
		Conn:           nc,
		Handler:        mux,
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

func handleFunc(prefix string) peanats.HandlerFunc {
	return func(pub peanats.Publisher, req peanats.Request) error {
		if ack, ok := pub.(peanats.Acker); ok {
			err := ack.Ack(nil)
			if err != nil {
				return err
			}
		}
		buf := new(bytes.Buffer)
		buf.WriteString(prefix)
		buf.WriteString(": ")
		buf.Write(req.Data())
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		return pub.Publish(buf.Bytes())
	}
}
