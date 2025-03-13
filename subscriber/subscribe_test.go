package subscriber_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/pond"
	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/subscriber"
)

func TestSubscribeChan(t *testing.T) {
	type Model struct{}
	t.Run("happy path", func(t *testing.T) {
		h := peanats.MsgHandlerFromArgHandler(peanats.ArgHandlerFunc[Model](
			func(ctx context.Context, a peanats.Arg[Model]) error {
				return nil
			},
		))
		ch, err := subscriber.SubscribeChan(t.Context(), h)
		require.NoError(t, err)
		close(ch)
		time.Sleep(time.Millisecond * 10)
	})
}

func BenchmarkSubscribeChan(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkSubscribeChan(b, subscriber.SubscribeBuffer(uint(b.N)))
	})
	b.Run("worker pool", func(b *testing.B) {
		benchmarkSubscribeChan(b, subscriber.SubscribeBuffer(uint(b.N)), subscriber.SubscribeSubmitter(pond.Submitter(250)))
	})
}

type benchmarkSubscribeArgument struct {
	Seq       int    `json:"seq"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type benchmarkSubscribeMsgImp struct {
	subject string
	data    []byte
}

func (m benchmarkSubscribeMsgImp) Subject() string {
	return m.subject
}

func (m benchmarkSubscribeMsgImp) Header() peanats.Header {
	return make(peanats.Header)
}

func (m benchmarkSubscribeMsgImp) Data() []byte {
	return m.data
}

func benchmarkSubscribeChan(b *testing.B, opts ...subscriber.SubscribeOption) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	subj := fmt.Sprintf("bench.subscribe-%d", b.N)

	wg := sync.WaitGroup{}
	wg.Add(b.N)

	h := peanats.ArgHandlerFunc[benchmarkSubscribeArgument](
		func(ctx context.Context, m peanats.Arg[benchmarkSubscribeArgument]) error {
			wg.Done()
			return nil
		},
	)

	ch, err := subscriber.SubscribeChan(b.Context(), peanats.MsgHandlerFromArgHandler(h), opts...)
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
		msg := benchmarkSubscribeMsgImp{
			subject: subj,
			data:    []byte(fmt.Sprintf(`{"seq":%d,"subject":"parson","predicate":"had","object":"a dog"}`, i)),
		}
		err := nc.Publish(b.Context(), msg)
		if err != nil {
			b.Fatal(err)
		}
	}
	wg.Wait()
	b.ReportAllocs()
}
