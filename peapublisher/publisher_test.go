package peapublisher_test

import (
	"fmt"
	"testing"

	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/peapublisher"
	"github.com/mikluko/peanats/xenc"
)

type testPayload struct {
	Seq int
	Str string
}

func BenchmarkPublisher(b *testing.B) {
	b.Run("single/json", func(b *testing.B) { benchmarkPublisher(b, peapublisher.WithContentType(xenc.ContentTypeJson)) })
	b.Run("pool/msgpack", func(b *testing.B) { benchmarkPublisher(b, peapublisher.WithContentType(xenc.ContentTypeMsgpack)) })
}

func benchmarkPublisher(b *testing.B, opts ...peapublisher.PublishOption) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	p := peapublisher.New(nc)
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
