package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/stream"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	c := stream.NewClient(peanats.NATS(nc))
	i := 0

	for _ = range time.Tick(time.Second) {
		rcv, err := c.Start(context.Background(), "peanats.stream", []byte(fmt.Sprintf("request: %d", i)))
		if err != nil {
			panic(err)
		}
		for {
			msg, err := rcv.Receive(context.Background())
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				panic(err)
			}
			log.Println(string(msg.Data))
		}
		i++
	}
}
