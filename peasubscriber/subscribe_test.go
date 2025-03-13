package peasubscriber_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats/contrib/pond"
	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/peasubscriber"
	"github.com/mikluko/peanats/xarg"
	"github.com/mikluko/peanats/xmsg"
)

func TestSubscribeChan(t *testing.T) {
	type Model struct{}
	t.Run("happy path", func(t *testing.T) {
		h := xarg.ArgMsgHandler(xarg.ArgHandlerFunc[Model](
			func(ctx context.Context, a xarg.Arg[Model]) error {
				return nil
			},
		))
		ch, err := peasubscriber.SubscribeChan(t.Context(), h)
		require.NoError(t, err)
		close(ch)
		time.Sleep(time.Millisecond * 10)
	})
}

func BenchmarkSubscribeChan(b *testing.B) {
	b.Run("single", func(b *testing.B) {
		benchmarkSubscribeChan(b, peasubscriber.SubscribeBuffer(uint(b.N)))
	})
	b.Run("worker pool", func(b *testing.B) {
		benchmarkSubscribeChan(b, peasubscriber.SubscribeBuffer(uint(b.N)), peasubscriber.SubscribeSubmitter(pond.Submitter(250)))
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

func (m benchmarkSubscribeMsgImp) Header() xmsg.Header {
	return make(xmsg.Header)
}

func (m benchmarkSubscribeMsgImp) Data() []byte {
	return m.data
}

func benchmarkSubscribeChan(b *testing.B, opts ...peasubscriber.SubscribeOption) {
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	subj := fmt.Sprintf("bench.subscribe-%d", b.N)

	wg := sync.WaitGroup{}
	wg.Add(b.N)

	h := xarg.ArgHandlerFunc[benchmarkSubscribeArgument](
		func(ctx context.Context, m xarg.Arg[benchmarkSubscribeArgument]) error {
			wg.Done()
			return nil
		},
	)

	ch, err := peasubscriber.SubscribeChan(b.Context(), xarg.ArgMsgHandler(h), opts...)
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
