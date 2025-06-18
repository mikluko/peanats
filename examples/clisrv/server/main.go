package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/logging"
	"github.com/mikluko/peanats/subscriber"
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

	h := peanats.ChainMsgMiddleware(
		peanats.MsgHandlerFromArgHandler[request](&handler{}),
		logging.AccessLogMiddleware(logging.SlogLogger(slog.Default(), slog.LevelInfo)),
	)
	ch, err := subscriber.SubscribeChan(ctx, h)
	if err != nil {
		panic(err)
	}
	sub, err := conn.SubscribeChan(ctx, "peanuts.examples.clisrv", ch)
	defer sub.Unsubscribe()

	<-ctx.Done()
}

type handler struct{}

func (handler) HandleArg(ctx context.Context, arg peanats.Arg[request]) error {
	x := arg.Value()
	slog.InfoContext(ctx, "received", "seq", x.Seq, "request", x.Request)
	return arg.(peanats.Respondable).Respond(ctx, &response{
		Seq:      x.Seq,
		Response: "response to " + x.Request,
	})
}
