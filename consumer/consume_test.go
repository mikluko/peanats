package consumer_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/consumer"
	"github.com/mikluko/peanats/contrib/pond"
	"github.com/mikluko/peanats/internal/xtestutil"
)

func BenchmarkConsume(b *testing.B) {
	b.Run("single/json", func(b *testing.B) {
		benchmarkConsume(b, xtestutil.Must(peanats.CodecContentType(peanats.ContentTypeJson)))
	})
	b.Run("single/msgpack", func(b *testing.B) {
		benchmarkConsume(b, xtestutil.Must(peanats.CodecContentType(peanats.ContentTypeMsgpack)))
	})
	// worker pool is expected to be slower vs single-threaded on no-op benchmark handler
	b.Run("pool/json", func(b *testing.B) {
		benchmarkConsume(b, xtestutil.Must(peanats.CodecContentType(peanats.ContentTypeJson)),
			consumer.ConsumeSubmitter(pond.Submitter(250)))
	})
	b.Run("pool/msgpack", func(b *testing.B) {
		benchmarkConsume(b, xtestutil.Must(peanats.CodecContentType(peanats.ContentTypeMsgpack)),
			consumer.ConsumeSubmitter(pond.Submitter(250)))
	})
}

type benchmarkConsumeArgument struct {
	Seq       int    `json:"seq"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

func benchmarkConsume(b *testing.B, codec peanats.Codec, opts ...consumer.ConsumeOption) {
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
		x := benchmarkConsumeArgument{
			Seq:       i,
			Subject:   "parson",
			Predicate: "had",
			Object:    "a dog",
		}
		p, err := codec.Marshal(x)
		if err != nil {
			b.Fatal(err)
		}
		h := make(peanats.Header)
		codec.SetContentType(h)
		err = nc.PublishMsg(&nats.Msg{
			Subject: subj,
			Header:  nats.Header(h),
			Data:    p,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	h := peanats.ArgHandlerFunc[benchmarkConsumeArgument](
		func(ctx context.Context, a peanats.Arg[benchmarkConsumeArgument]) error {
			defer wg.Done()
			return a.(peanats.Ackable).Ack(ctx)
		},
	)

	b.ResetTimer()

	err = consumer.Consume(b.Context(), c, peanats.MsgHandlerFromArgHandler(h), opts...)
	if err != nil {
		b.Fatal(err)
	}

	wg.Wait()
	b.ReportAllocs()
}
