package requester

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
	"github.com/mikluko/peanats/subscriber"
)

func TestRequester_Request(t *testing.T) {
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
		c := New[request, response](nc)
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
		c := New[request, response](nc)
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
		c := New[request, response](nc)
		rs, err := c.Request(context.Background(), "parson.had", &request{Foo: "a dog"})
		require.Error(t, err)
		assert.ErrorIs(t, err, testErr)
		assert.Nil(t, rs)
	})
}

func BenchmarkRequester_Request(b *testing.B) {
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

	c := New[request, response](nc)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := c.Request(b.Context(), "baz.qux", &request{Foo: "foo"})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRequester_ResponseReceiver(b *testing.B) {
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

	c := New[request, response](nc)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		n := i % 10
		rcv, err := c.ResponseReceiver(b.Context(), "baz.qux", &request{N: n}, ResponseReceiverBuffer(10))
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
		if !errors.Is(err, ErrOver) {
			b.Fatal(err)
		}
		if !errors.Is(err, io.EOF) {
			b.Fatal(err)
		}
	}
}

func TestRequestHeader_Merging(t *testing.T) {
	t.Run("merging", func(t *testing.T) {
		// Test that RequestHeader merges headers instead of replacing them

		// Start with default params (includes Content-Type: application/json)
		params := makeRequestParams()

		// Verify default Content-Type is set
		assert.Equal(t, []string{peanats.ContentTypeJson.String()}, params.header[peanats.HeaderContentType])

		// Add custom headers using RequestHeader
		customHeader := make(peanats.Header)
		customHeader.Set("Authorization", "Bearer token123")
		customHeader.Set("X-Custom", "value1")

		RequestHeader(customHeader)(&params)

		// Verify both default and custom headers are present
		assert.Equal(t, []string{peanats.ContentTypeJson.String()}, params.header[peanats.HeaderContentType])
		assert.Equal(t, []string{"Bearer token123"}, params.header["Authorization"])
		assert.Equal(t, []string{"value1"}, params.header["X-Custom"])

		// Add more headers to the same key
		moreHeaders := make(peanats.Header)
		moreHeaders.Add("X-Custom", "value2")
		moreHeaders.Set("X-Another", "another")

		RequestHeader(moreHeaders)(&params)

		// Verify headers are merged/appended
		assert.Equal(t, []string{peanats.ContentTypeJson.String()}, params.header[peanats.HeaderContentType])
		assert.Equal(t, []string{"Bearer token123"}, params.header["Authorization"])
		assert.Equal(t, []string{"value1", "value2"}, params.header["X-Custom"]) // Both values present
		assert.Equal(t, []string{"another"}, params.header["X-Another"])
	})
	t.Run("multiple options", func(t *testing.T) {
		// Test using multiple RequestHeader options in sequence

		header1 := make(peanats.Header)
		header1.Set("X-First", "first")

		header2 := make(peanats.Header)
		header2.Set("X-Second", "second")

		header3 := make(peanats.Header)
		header3.Set("X-First", "first-updated") // Same key, should append

		params := makeRequestParams(
			RequestHeader(header1),
			RequestHeader(header2),
			RequestHeader(header3),
		)

		// Verify all headers are present
		assert.Equal(t, []string{peanats.ContentTypeJson.String()}, params.header[peanats.HeaderContentType])
		assert.Equal(t, []string{"first", "first-updated"}, params.header["X-First"])
		assert.Equal(t, []string{"second"}, params.header["X-Second"])
	})
	t.Run("override", func(t *testing.T) {
		params := makeRequestParams(
			RequestContentType(peanats.ContentTypeYaml),
		)
		assert.Equal(t, []string{peanats.ContentTypeYaml.String()}, params.header[peanats.HeaderContentType])
	})
	t.Run("with content type", func(t *testing.T) {
		customHeader := make(peanats.Header)
		customHeader.Set("Authorization", "Bearer token123")

		params := makeRequestParams(
			RequestHeader(customHeader),
			RequestContentType(peanats.ContentTypeYaml), // This should override default JSON
		)

		assert.Equal(t, []string{peanats.ContentTypeYaml.String()}, params.header[peanats.HeaderContentType])
		assert.Equal(t, []string{"Bearer token123"}, params.header["Authorization"])
	})
}
