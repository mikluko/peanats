package main

import (
	"context"
	"log/slog"
	"os"

	slogcontrib "github.com/mikluko/peanats/contrib/slog"
	"github.com/mikluko/peanats/contrib/xmw"
	"github.com/mikluko/peanats/peasubscriber"
	"github.com/mikluko/peanats/xarg"
	"github.com/mikluko/peanats/xmsg"
	"github.com/mikluko/peanats/xnats"

	"os/signal"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := xnats.Wrap(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	h := xmsg.ChainMsgMiddleware(
		xarg.ArgMsgHandler(xarg.ArgHandlerFunc[model](handleModel)),
		xmw.AccessLogMiddleware(xmw.WithAccessLogMiddlewareLogger(slogcontrib.Logger(slog.Default(), slog.LevelInfo))),
	)
	ch, err := peasubscriber.SubscribeChan(ctx, h)
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

func handleModel(_ context.Context, arg xarg.Arg[model]) error {
	x := arg.Value()
	slog.Info(x.Msg, "seq", x.Seq)
	return nil
}
