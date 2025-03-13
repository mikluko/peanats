package peanats_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/pond"
	"github.com/mikluko/peanats/internal/xtestutil"
)

func BenchmarkSubscribe(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkSubscribeChan(b, peanats.SubscribeBuffer(uint(b.N)))
	})
	b.Run("worker pool", func(b *testing.B) {
		benchmarkSubscribeChan(b, peanats.SubscribeBuffer(uint(b.N)), peanats.SubscribeSubmitter(pond.Submitter(250)))
	})
}

type benchmarkSubscribeArgument struct {
	Seq       int    `json:"seq"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

func benchmarkSubscribeChan(b *testing.B, opts ...peanats.SubscribeOption) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	subj := fmt.Sprintf("bench.subscribe-%d", b.N)

	wg := sync.WaitGroup{}
	wg.Add(b.N)

	h := peanats.ArgumentHandlerFunc[benchmarkSubscribeArgument](
		func(ctx context.Context, m peanats.Argument[benchmarkSubscribeArgument]) error {
			wg.Done()
			return nil
		},
	)

	ch, err := peanats.SubscribeChan(b.Context(), peanats.ArgumentMessageHandler(h), opts...)
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
