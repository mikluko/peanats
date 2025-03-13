package main

import (
	"context"
	"log/slog"
	"os"

	"os/signal"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/logging"
	slogcontrib "github.com/mikluko/peanats/contrib/slog"
	"github.com/mikluko/peanats/subscriber"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := peanats.WrapConnection(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	h := peanats.ChainMsgMiddleware(
		peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[model](handleModel)),
		logging.AccessLogMiddleware(logging.AccessLogMiddlewareLogger(slogcontrib.Logger(slog.Default(), slog.LevelInfo))),
	)
	ch, err := subscriber.SubscribeChan(ctx, h)
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

func handleModel(_ context.Context, arg peanats.Arg[model]) error {
	x := arg.Value()
	slog.Info(x.Msg, "seq", x.Seq)
	return nil
}
