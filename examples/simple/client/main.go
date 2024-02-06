package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return requester(ctx, nc)
	})
	g.Go(func() error {
		return observer(ctx, nc)
	})
	err = g.Wait()
	if err != nil {
		panic(err)
	}
}

func requester(ctx context.Context, nc *nats.Conn) error {
	t := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			err := nc.Publish("peanuts.simple.requests", []byte(time.Now().String()))
			if err != nil {
				return err
			}
		}
	}
}

func observer(ctx context.Context, nc *nats.Conn) error {
	ch := make(chan *nats.Msg)
	sub, err := nc.ChanSubscribe("peanuts.simple.results", ch)
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
			log.Println(seq, string(msg.Data))
		}
	}
}
