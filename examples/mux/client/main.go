package main

import (
	"context"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"time"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	requesterA := requesterFunc("request A: ", "peanuts.mux.requests.a")
	g.Go(func() error {
		return requesterA(ctx, nc)
	})
	requesterB := requesterFunc("request B: ", "peanuts.mux.requests.b")
	g.Go(func() error {
		return requesterB(ctx, nc)
	})
	g.Go(func() error {
		return observer(ctx, nc)
	})
	err = g.Wait()
	if err != nil {
		panic(err)
	}
}

func requesterFunc(prefix, subj string) func(ctx context.Context, nc *nats.Conn) error {
	return func(ctx context.Context, nc *nats.Conn) error {
		t := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-t.C:
				_, err := nc.RequestWithContext(ctx, subj, []byte(prefix+time.Now().String()))
				if err != nil {
					return err
				}
			}
		}
	}
}

func observer(ctx context.Context, nc *nats.Conn) error {
	ch := make(chan *nats.Msg, 1)
	sub, err := nc.ChanSubscribe("peanuts.mux.results.>", ch)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	seq := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			seq++
			println(seq, string(msg.Data))
		}
	}
}
