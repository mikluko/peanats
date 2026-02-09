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

	conn, err := transport.Wrap(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}

	disp := peanats.NewDispatcher()

	h := peanats.ChainMsgMiddleware(
		peanats.MsgHandlerFromArgHandler[request](&handler{}),
		logging.AccessLogMiddleware(logging.SlogLogger(slog.Default(), slog.LevelInfo)),
	)
	ch, err := subscriber.SubscribeChan(ctx, h, subscriber.SubscribeDispatcher(disp))
	if err != nil {
		panic(err)
	}
	sub, err := conn.SubscribeChan(ctx, "peanuts.examples.clisrv", ch)
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

type handler struct{}

func (handler) HandleArg(ctx context.Context, arg peanats.Arg[request]) error {
	x := arg.Value()
	slog.InfoContext(ctx, "received", "seq", x.Seq, "request", x.Request)
	return arg.(peanats.Respondable).Respond(ctx, &response{
		Seq:      x.Seq,
		Response: "response to " + x.Request,
	})
}
