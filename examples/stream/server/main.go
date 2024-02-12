package main

import (
	"fmt"
	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/stream"
	"github.com/nats-io/nats.go"
	"math/rand"
	"sync/atomic"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	count := int64(0)
	srv := peanats.Server{
		Conn:           peanats.NATS(nc),
		ListenSubjects: []string{"peanats.stream"},
		Handler: peanats.ChainMiddleware(
			peanats.HandlerFunc(func(pub peanats.Publisher, req peanats.Request) error {
				seq := atomic.AddInt64(&count, 1)
				for i := int64(0); i < rand.Int63n(10); i++ {
					err := pub.Publish([]byte(fmt.Sprintf("%s, response: %d", req.Data(), seq)))
					if err != nil {
						return err
					}
					seq++
				}
				return nil
			}),
			stream.Middleware,
			peanats.MakeAccessLogMiddleware(),
		),
	}
	err = srv.Start()
	if err != nil {
		panic(err)
	}
	srv.Wait()
}
