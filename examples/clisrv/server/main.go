package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"

	slogcontrib "github.com/mikluko/peanats/contrib/slog"
	"github.com/mikluko/peanats/contrib/xmw"
	"github.com/mikluko/peanats/peasubscriber"
	"github.com/mikluko/peanats/xarg"
	"github.com/mikluko/peanats/xmsg"
	"github.com/mikluko/peanats/xnats"
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

	conn, err := xnats.Wrap(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}

	h := xmsg.ChainMsgMiddleware(
		xarg.ArgMsgHandler[request](&handler{}),
		xmw.AccessLogMiddleware(xmw.WithAccessLogMiddlewareLogger(slogcontrib.Logger(slog.Default(), slog.LevelInfo))),
	)
	ch, err := peasubscriber.SubscribeChan(ctx, h)
	if err != nil {
		panic(err)
	}
	sub, err := conn.ChanSubscribe("peanuts.examples.clisrv", ch)
	defer sub.Unsubscribe()

	<-ctx.Done()
}

type handler struct{}

func (handler) HandleArg(ctx context.Context, arg xarg.Arg[request]) error {
	x := arg.Value()
	slog.InfoContext(ctx, "received", "seq", x.Seq, "request", x.Request)
	return arg.(xmsg.Respondable).Respond(ctx, &response{
		Seq:      x.Seq,
		Response: "response to " + x.Request,
	})
}
