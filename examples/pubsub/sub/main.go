package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/mikluko/peanats"
	slogcontrib "github.com/mikluko/peanats/contrib/slog"

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

	h := peanats.ChainMessageMiddleware(
		peanats.ArgumentMessageHandler(peanats.ArgumentHandlerFunc[model](handleModel)),
		peanats.AccessLogMiddleware(peanats.WithAccessLogMiddlewareLogger(slogcontrib.Logger(slog.Default(), slog.LevelInfo))),
	)
	ch, err := peanats.SubscribeChan(ctx, h)
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

func handleModel(_ context.Context, arg peanats.Argument[model]) error {
	x := arg.Value()
	slog.Info(x.Msg, "seq", x.Seq)
	return nil
}
