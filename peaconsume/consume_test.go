package peaconsume

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/alitto/pond/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xtestutil"
)

func BenchmarkConsume(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkConsume(b)
	})
	// worker pool is expected to be slower vs single-threaded on no-op benchmark handler
	b.Run("worker pool", func(b *testing.B) {
		pool := pond.NewPool(250)
		benchmarkConsume(b, WithExecutor(func(f func()) {
			pool.Submit(f)
		}))
	})
}

type benchmarkArgument struct {
	Seq       int    `json:"seq"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

func benchmarkConsume(b *testing.B, opts ...Option) {
	ns := xtestutil.Server(b)

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		b.Fatal(err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		b.Fatal(err)
	}

	name := fmt.Sprintf("bench-%d", b.N)
	subj := fmt.Sprintf("bench.consume.%d", b.N)

	s, err := js.CreateStream(b.Context(), jetstream.StreamConfig{
		Name:     name,
		Subjects: []string{subj},
	})
	if err != nil {
		b.Fatal(err)
	}

	c, err := s.CreateConsumer(b.Context(), jetstream.ConsumerConfig{
		Name:    name,
		Durable: name,
	})
	if err != nil {
		b.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		x := benchmarkArgument{
			Seq:       i,
			Subject:   "parson",
			Predicate: "had",
			Object:    "a dog",
		}
		p, err := json.Marshal(x)
		if err != nil {
			b.Fatal(err)
		}
		err = nc.Publish(subj, p)
		if err != nil {
			b.Fatal(err)
		}
	}

	h := peanats.ArgumentHandlerFunc[benchmarkArgument](
		func(ctx context.Context, d peanats.Dispatcher, a peanats.Argument[benchmarkArgument]) {
			if err := d.(Dispatcher).Ack(ctx); err != nil {
				b.Error(err)
			}
			wg.Done()
		},
	)

	b.ResetTimer()

	err = Consume(b.Context(), c, Handler(h), opts...)
	if err != nil {
		b.Fatal(err)
	}

	wg.Wait()
	b.ReportAllocs()
}
