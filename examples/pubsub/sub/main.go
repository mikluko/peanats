package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/logging"
	"github.com/mikluko/peanats/subscriber"
	"github.com/mikluko/peanats/transport"
)

func main() {
	nc, err := transport.Wrap(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	disp := peanats.NewDispatcher()

	h := peanats.ChainMsgMiddleware(
		peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[model](handleModel)),
		logging.AccessLogMiddleware(logging.SlogLogger(slog.Default(), slog.LevelInfo)),
	)
	ch, err := subscriber.SubscribeChan(ctx, h, subscriber.SubscribeDispatcher(disp))
	if err != nil {
		panic(err)
	}

	sub, err := nc.SubscribeChan(ctx, "peanuts.examples.pubsub", ch)
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	<-ctx.Done()

	// Wait for in-flight handlers to complete and collect errors.
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer waitCancel()
	if err := disp.Wait(waitCtx); err != nil {
		slog.Error("dispatcher drain failed", "error", err)
	}
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
