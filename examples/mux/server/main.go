package main

import (
	"bytes"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	mux := new(peanats.ServeMux)
	_ = mux.Handle(
		peanats.ChainMiddleware(handleFunc("handle A:"), peanats.MakePublishSubjectMiddleware("peanuts.mux.results.a")),
		"peanuts.mux.requests.a",
	)
	_ = mux.Handle(
		peanats.ChainMiddleware(handleFunc("handle B:"), peanats.MakePublishSubjectMiddleware("peanuts.mux.results.b")),
		"peanuts.mux.requests.b",
	)
	srv := peanats.Server{
		ListenSubjects: []string{"peanuts.mux.requests.>"},
		Conn:           peanats.NATS(nc),
		Handler: peanats.ChainMiddleware(mux,
			peanats.MakeAccessLogMiddleware(),
		),
	}

	err = srv.Start()
	if err != nil {
		panic(err)
	}

	srv.Wait()
}

func handleFunc(prefix string) peanats.HandlerFunc {
	return func(pub peanats.Publisher, req peanats.Request) error {
		buf := new(bytes.Buffer)
		buf.WriteString(prefix)
		buf.WriteString(": ")
		buf.Write(req.Data())
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		return pub.Publish(buf.Bytes())
	}
}
