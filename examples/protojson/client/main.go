package main

import (
	"context"
	"fmt"
	"github.com/mikluko/peanats/examples/protojson/api"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
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
			req := api.Argument{Arg: time.Now().String()}
			data, err := protojson.Marshal(&req)
			if err != nil {
				panic(err)
			}
			_, err = nc.RequestWithContext(ctx, "peanuts.protojson.requests", data)
			if err != nil {
				return err
			}
		}
	}
}

func observer(ctx context.Context, nc *nats.Conn) error {
	ch := make(chan *nats.Msg)
	sub, err := nc.ChanSubscribe("peanuts.protojson.results", ch)
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
			res := new(api.Result)
			err = protojson.Unmarshal(msg.Data, res)
			if err != nil {
				panic(err)
			}
			fmt.Println(seq, res.String())
		}
	}
}
