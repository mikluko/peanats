package requester_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/contrib/pond"
	"github.com/mikluko/peanats/internal/xmock/peanatsmock"
	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/requester"
	"github.com/mikluko/peanats/subscriber"
)

func TestClientImpl_Request(t *testing.T) {
	type request struct {
		Foo string `json:"foo"`
	}
	type response struct {
		Bar string `json:"bar"`
	}
	t.Run("happy path", func(t *testing.T) {
		msg := peanatsmock.NewMsg(t)
		msg.EXPECT().Data().Return([]byte(`{"bar": "a dog"}`)).Once()
		msg.EXPECT().Header().Return(peanats.Header{})
		nc := peanatsmock.NewConnection(t)
		nc.EXPECT().
			Request(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg peanats.Msg) {
				assert.Equal(t, "parson.had", msg.Subject())
				assert.JSONEq(t, `{"foo": "a dog"}`, string(msg.Data()))
			}).
			Return(msg, nil).Once()
		c := requester.New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.NoError(t, err)
		require.NotNil(t, rs)
		assert.Equal(t, &response{Bar: "a dog"}, rs.Value())
	})
	t.Run("decode error", func(t *testing.T) {
		msg := peanatsmock.NewMsg(t)
		msg.EXPECT().Data().Return([]byte(`{`)).Once()
		msg.EXPECT().Header().Return(peanats.Header{})
		nc := peanatsmock.NewConnection(t)
		nc.EXPECT().
			Request(mock.Anything, mock.Anything).
			Return(msg, nil).Once()
		c := requester.New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		require.Nil(t, rs)
	})
	t.Run("some error", func(t *testing.T) {
		testErr := errors.New("parson had a dog")
		nc := peanatsmock.NewConnection(t)
		nc.EXPECT().
			Request(mock.Anything, mock.Anything).
			Return(nil, testErr).Once()
		c := requester.New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.Nil(t, rs)
	})
}

func BenchmarkClientImpl_Request(b *testing.B) {
	type request struct {
		Foo string `json:"foo"`
	}
	type response struct {
		Bar string `json:"bar"`
	}
	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	argh := peanats.ArgHandlerFunc[request](func(ctx context.Context, arg peanats.Arg[request]) error {
		return arg.(peanats.Respondable).Respond(ctx, &response{Bar: arg.Value().Foo})
	})
	msgh := peanats.MsgHandlerFromArgHandler(argh)
	sub, err := nc.SubscribeHandler(b.Context(), "baz.qux", msgh)

	if err != nil {
		b.Fatal(err)
	}
	defer sub.Unsubscribe()

	c := requester.New[request, response](nc)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := c.Request(b.Context(), "baz.qux", &request{Foo: "foo"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClientImpl_ResponseReceiver(b *testing.B) {
	type request struct {
		N int `json:"n"`
	}
	type response struct {
		Seq int `json:"seq"`
	}

	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)

	argh := peanats.ArgHandlerFunc[request](func(ctx context.Context, arg peanats.Arg[request]) error {
		n := arg.Value().N % 10
		for i := 0; i < n; i++ {
			err := arg.(peanats.Respondable).Respond(ctx, &response{Seq: i})
			if err != nil {
				return err
			}
		}
		return arg.(peanats.Respondable).Respond(ctx, nil)
	})
	subch, _ := subscriber.SubscribeChan(
		b.Context(),
		peanats.MsgHandlerFromArgHandler(argh),
		subscriber.SubscribeSubmitter(pond.Submitter(1000)),
	)
	sub, err := nc.SubscribeChan(b.Context(), "baz.qux", subch)

	if err != nil {
		b.Fatal(err)
	}
	defer sub.Unsubscribe()

	c := requester.New[request, response](nc)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		n := i % 10
		rcv, err := c.ResponseReceiver(b.Context(), "baz.qux", &request{N: n}, requester.ResponseReceiverBuffer(10))
		if err != nil {
			b.Fatal(err)
		}
		for j := 0; j < n; j++ {
			_, err := rcv.Next(b.Context())
			if err != nil {
				b.Fatal(err)
			}
		}
		_, err = rcv.Next(b.Context())
		if !errors.Is(err, requester.ErrOver) {
			b.Fatal(err)
		}
		if !errors.Is(err, io.EOF) {
			b.Fatal(err)
		}
	}
}
