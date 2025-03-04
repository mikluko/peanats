package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/peasubscribe"

	"os/signal"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	handler := peasubscribe.Handler(peanats.ArgumentHandlerFunc[model](handleModel))
	ch, err := peasubscribe.SubscribeChan(ctx, handler)
	if err != nil {
		panic(err)
	}

	sub, err := nc.ChanSubscribe("peanuts.examples.pubsub", ch)
	defer sub.Unsubscribe()

	<-ctx.Done()
}

type model struct {
	Seq uint   `json:"seq"`
	Msg string `json:"msg"`
}

func handleModel(ctx context.Context, d peanats.Dispatcher, a peanats.Argument[model]) {
	obj := a.Payload()
	slog.InfoContext(ctx, "received", "seq", obj.Seq, "msg", obj.Msg)
}
