package peanats_test

import (
	"fmt"
	"testing"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/internal/xtestutil"
)

type testPayload struct {
	Seq int
	Str string
}

func BenchmarkPublisher(b *testing.B) {
	b.Run("single/json", func(b *testing.B) { benchmarkPublisher(b, peanats.PublishContentType(peanats.ContentTypeJson)) })
	b.Run("pool/msgpack", func(b *testing.B) { benchmarkPublisher(b, peanats.PublishContentType(peanats.ContentTypeMsgpack)) })
}

func benchmarkPublisher(b *testing.B, opts ...peanats.PublishOption) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	p := peanats.NewPublisher(nc)
	subj := fmt.Sprintf("bench.publisher-%d", b.N)

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
