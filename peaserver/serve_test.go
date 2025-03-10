package peaserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/alitto/pond/v2"
	"github.com/nats-io/nats.go"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xtestutil"
)

func BenchmarkServe(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkServe(b)
	})
	// worker pool is expected to be slower vs single-threaded on no-op benchmark handler
	b.Run("worker pool", func(b *testing.B) {
		pool := pond.NewPool(250)
		benchmarkServe(b, WithExecutor(func(f func()) {
			pool.Submit(f)
		}))
	})
}

type benchmarkRequest struct {
	Seq       int    `json:"seq"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
}

type benchmarkResponse struct {
	Seq    int    `json:"seq"`
	Object string `json:"object"`
}

func benchmarkServe(b *testing.B, opts ...Option) {
	ns := xtestutil.Server(b)

	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		b.Fatal(err)
	}
	defer nc.Close()

	wg := sync.WaitGroup{}
	wg.Add(b.N)

	h := peanats.ArgumentHandlerFunc[benchmarkRequest](func(ctx context.Context, d peanats.Dispatcher, arg peanats.Argument[benchmarkRequest]) {
		rq := arg.Payload()
		err := d.(Dispatcher[benchmarkResponse]).Respond(ctx, &benchmarkResponse{
			Seq: rq.Seq,
		})
		if err != nil {
			d.Error(ctx, err)
		}
		wg.Done()
	})

	subj := fmt.Sprintf("bench.serve.%d", b.N)

	ch, err := ServeChan(b.Context(), Handler[benchmarkRequest, benchmarkResponse](h), opts...)
	sub, err := nc.ChanSubscribe(subj, ch)
	b.Cleanup(func() {
		sub.Unsubscribe()
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := benchmarkRequest{
			Seq:       i,
			Subject:   "parson",
			Predicate: "had",
		}
		p, err := json.Marshal(x)
		if err != nil {
			b.Fatal(err)
		}
		_, err = nc.RequestWithContext(b.Context(), subj, p)
		if err != nil {
			b.Fatal(err)
		}
	}

	wg.Wait()
	b.ReportAllocs()
}
