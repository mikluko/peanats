package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
	"github.com/mikluko/peanats/publisher"
	"github.com/mikluko/peanats/transport"
)

type message struct {
	subject string
	Seq     uint   `json:"seq"`
	Msg     string `json:"msg"`
}

func (m *message) Subject() string {
	return m.subject
}

func (m *message) Data() []byte {
	p, _ := json.Marshal(m)
	return p
}

func (m *message) Header() peanats.Header {
	return peanats.Header{codec.HeaderContentType: []string{"application/json"}}
}

func main() {
	nc, err := transport.Wrap(nats.Connect(nats.DefaultURL))
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pub := publisher.New(nc)
	seq := uint(0)

	for range time.Tick(1 * time.Second) {
		seq++
		msg := message{subject: "peanuts.examples.pubsub", Seq: seq, Msg: "parson had a dog"}
		err = pub.Publish(ctx, "peanuts.examples.pubsub", msg)
		if err != nil {
			panic(err)
		}
	}
}
