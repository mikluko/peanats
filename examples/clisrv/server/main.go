package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/peaserver"
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

	h := peanats.ChainMiddleware(
		peaserver.Handler[request, response](&handler{}),
		peanats.AccessLogMiddleware(peanats.WithAccessLogMiddlewareLogger(peanats.NewSlogLogger(slog.Default(), slog.LevelInfo))),
	)
	ch, err := peaserver.ServeChan(ctx, h)
	if err != nil {
		panic(err)
	}
	sub, err := nc.ChanSubscribe("peanuts.examples.clisrv", ch)
	defer sub.Unsubscribe()

	<-ctx.Done()
}

type handler struct{}

func (handler) Handle(ctx context.Context, d peanats.Dispatcher, a peanats.Argument[request]) {
	rq := a.Payload()
	dd := d.(peaserver.Dispatcher[response])
	err := dd.Respond(ctx, &response{
		Seq:      rq.Seq,
		Response: "response to " + rq.Request,
	})
	if err != nil {
		d.Error(ctx, err)
	}
	slog.InfoContext(ctx, "received", "seq", rq.Seq, "request", rq.Request)
}
