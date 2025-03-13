package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
)

type model struct {
	Seq uint   `json:"seq"`
	Msg string `json:"msg"`
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pub := peanats.NewPublisher(nc)
	seq := uint(0)

	for _ = range time.Tick(1 * time.Second) {
		seq++
		obj := model{Seq: seq, Msg: "parson had a dog"}
		err = pub.Publish(ctx, "peanuts.examples.pubsub", &obj)
		if err != nil {
			panic(err)
		}
	}
}
