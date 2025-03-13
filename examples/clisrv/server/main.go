package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	slogcontrib "github.com/mikluko/peanats/contrib/slog"
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

	h := peanats.ChainMessageMiddleware(
		peanats.ArgumentMessageHandler[request](&handler{}),
		peanats.AccessLogMiddleware(peanats.WithAccessLogMiddlewareLogger(slogcontrib.Logger(slog.Default(), slog.LevelInfo))),
	)
	ch, err := peanats.SubscribeChan(ctx, h)
	if err != nil {
		panic(err)
	}
	sub, err := nc.ChanSubscribe("peanuts.examples.clisrv", ch)
	defer sub.Unsubscribe()

	<-ctx.Done()
}

type handler struct{}

func (handler) HandleArgument(ctx context.Context, a peanats.Argument[request]) error {
	rq := a.Value()
	slog.InfoContext(ctx, "received", "seq", rq.Seq, "request", rq.Request)
	return a.Respond(ctx, &response{
		Seq:      rq.Seq,
		Response: "response to " + rq.Request,
	})
}
