package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/requester"
)

type request struct {
	Seq     uint   `json:"seq"`
	Request string `json:"request"`
}

type response struct {
	Seq      uint   `json:"seq"`
	Response string `json:"response"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conn, err := peanats.WrapConnection(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}

	client := requester.New[request, response](conn)
	for t := range time.Tick(1 * time.Second) {
		rqCtx, rqCancel := context.WithTimeout(ctx, 100*time.Millisecond)
		req := request{Seq: 1, Request: t.Format(time.RFC3339)}
		res, err := client.Request(rqCtx, "peanuts.examples.clisrv", &req)
		if err != nil {
			panic(err)
		}
		x := res.Value()
		slog.Info("response", "seq", x.Seq, "response", x.Response)
		rqCancel()
	}
}
