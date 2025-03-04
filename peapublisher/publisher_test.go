package peapublisher

import (
	"fmt"
	"testing"

	"github.com/mikluko/peanats/internal/xtestutil"
)

type testPayload struct {
	Seq int
	Str string
}

func BenchmarkPublisher(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkPublisher(b)
	})
}

func benchmarkPublisher(b *testing.B) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	p := New(nc)
	subj := fmt.Sprintf("bench.publisher-%d", b.N)

	for i := 0; i < b.N; i++ {
		err := p.Publish(b.Context(), subj, &testPayload{
			Seq: i,
			Str: "parson had a dog",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
