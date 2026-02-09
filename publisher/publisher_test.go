package publisher_test

import (
	"fmt"
	"testing"

	"github.com/mikluko/peanats/codec"
	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/publisher"
)

type testPayload struct {
	Seq int
	Str string
}

func BenchmarkPublisher(b *testing.B) {
	b.Run("single/json", func(b *testing.B) { benchmarkPublisher(b, publisher.WithContentType(codec.JSON)) })
	b.Run("pool/msgpack", func(b *testing.B) { benchmarkPublisher(b, publisher.WithContentType(codec.Msgpack)) })
}

func benchmarkPublisher(b *testing.B, opts ...publisher.PublishOption) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	p := publisher.New(nc)
	subj := fmt.Sprintf("bench.publisher-%d", b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := p.Publish(b.Context(), subj, &testPayload{
			Seq: i,
			Str: "parson had a dog",
		}, opts...)
		if err != nil {
			b.Fatal(err)
		}
	}
}
