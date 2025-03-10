package peasubscriber

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/alitto/pond/v2"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xtestutil"
)

func BenchmarkSubscribe(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkSubscribeChan(b, WithBuffer(uint(b.N)))
	})
	b.Run("worker pool", func(b *testing.B) {
		pool := pond.NewPool(250)
		benchmarkSubscribeChan(b, WithBuffer(uint(b.N)), WithExecutor(func(f func()) {
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

func benchmarkSubscribeChan(b *testing.B, opts ...Option) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	subj := fmt.Sprintf("bench.subscribe-%d", b.N)

	wg := sync.WaitGroup{}
	wg.Add(b.N)

	h := peanats.ArgumentHandlerFunc[benchmarkArgument](
		func(ctx context.Context, d peanats.Dispatcher, m peanats.Argument[benchmarkArgument]) {
			wg.Done()
		},
	)

	ch, err := SubscribeChan(b.Context(), Handler(h), opts...)
	if err != nil {
		b.Fatal(err)
	}

	sub, err := nc.ChanSubscribe(subj, ch)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		sub.Unsubscribe()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := nc.Publish(subj, []byte(fmt.Sprintf(`{"seq":%d,"subject":"parson","predicate":"had","object":"a dog"}`, i)))
		if err != nil {
			b.Fatal(err)
		}
	}
	wg.Wait()
	b.ReportAllocs()
}
