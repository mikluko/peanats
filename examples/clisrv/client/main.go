package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats/peaclient"
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

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	client := peaclient.New[request, response](nc)
	for t := range time.Tick(1 * time.Second) {
		rqCtx, rqCancel := context.WithTimeout(ctx, 100*time.Millisecond)
		req := request{Seq: 1, Request: t.Format(time.RFC3339)}
		res, err := client.Request(rqCtx, "peanuts.examples.clisrv", &req)
		if err != nil {
			panic(err)
		}
		slog.Info("response", "seq", res.Payload().Seq, "response", res.Payload().Response)
		rqCancel()
	}
}
