package publisher_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
	"github.com/mikluko/peanats/internal/xmock/transportmock"
	"github.com/mikluko/peanats/internal/xtestutil"
	"github.com/mikluko/peanats/publisher"
)

type testPayload struct {
	Seq int    `json:"seq" yaml:"seq" msgpack:"seq"`
	Str string `json:"str" yaml:"str" msgpack:"str"`
}

func TestPublisher_Publish(t *testing.T) {
	t.Run("default json", func(t *testing.T) {
		nc := transportmock.NewConn(t)
		nc.EXPECT().
			Publish(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg peanats.Msg) {
				assert.Equal(t, "test.subject", msg.Subject())
				assert.Equal(t, codec.JSON.String(), msg.Header().Get(codec.HeaderContentType))
				assert.JSONEq(t, `{"seq":1,"str":"hello"}`, string(msg.Data()))
			}).
			Return(nil).Once()

		p := publisher.New(nc)
		err := p.Publish(context.Background(), "test.subject", &testPayload{Seq: 1, Str: "hello"})
		require.NoError(t, err)
	})

	t.Run("with content type", func(t *testing.T) {
		nc := transportmock.NewConn(t)
		nc.EXPECT().
			Publish(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg peanats.Msg) {
				assert.Equal(t, codec.YAML.String(), msg.Header().Get(codec.HeaderContentType))
				assert.Contains(t, string(msg.Data()), "seq: 1")
			}).
			Return(nil).Once()

		p := publisher.New(nc)
		err := p.Publish(context.Background(), "test.subject", &testPayload{Seq: 1, Str: "hello"},
			publisher.WithContentType(codec.YAML),
		)
		require.NoError(t, err)
	})

	t.Run("with header", func(t *testing.T) {
		nc := transportmock.NewConn(t)
		nc.EXPECT().
			Publish(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg peanats.Msg) {
				assert.Equal(t, "bar", msg.Header().Get("X-Foo"))
				assert.Equal(t, codec.JSON.String(), msg.Header().Get(codec.HeaderContentType))
			}).
			Return(nil).Once()

		p := publisher.New(nc)
		err := p.Publish(context.Background(), "test.subject", &testPayload{Seq: 1, Str: "hello"},
			publisher.WithHeader(peanats.Header{"X-Foo": []string{"bar"}}),
		)
		require.NoError(t, err)
	})

	t.Run("with content encoding", func(t *testing.T) {
		encodings := []struct {
			name     string
			encoding codec.ContentEncoding
		}{
			{"zstd", codec.Zstd},
			{"s2", codec.S2},
		}
		for _, enc := range encodings {
			t.Run(enc.name, func(t *testing.T) {
				var captured peanats.Msg
				nc := transportmock.NewConn(t)
				nc.EXPECT().
					Publish(mock.Anything, mock.Anything).
					Run(func(_ context.Context, msg peanats.Msg) {
						captured = msg
					}).
					Return(nil).Once()

				p := publisher.New(nc)
				err := p.Publish(context.Background(), "test.subject", &testPayload{Seq: 1, Str: "hello"},
					publisher.WithContentEncoding(enc.encoding),
				)
				require.NoError(t, err)

				// Verify encoding header is set
				assert.Equal(t, enc.encoding.String(), captured.Header().Get(codec.HeaderContentEncoding))

				// Verify data is compressed (not plain JSON)
				assert.NotEqual(t, `{"seq":1,"str":"hello"}`, string(captured.Data()))

				// Verify round-trip: decompress and unmarshal
				var result testPayload
				err = codec.UnmarshalHeader(captured.Data(), &result, captured.Header())
				require.NoError(t, err)
				assert.Equal(t, testPayload{Seq: 1, Str: "hello"}, result)
			})
		}
	})

	t.Run("nil payload", func(t *testing.T) {
		nc := transportmock.NewConn(t)
		nc.EXPECT().
			Publish(mock.Anything, mock.Anything).
			Run(func(_ context.Context, msg peanats.Msg) {
				assert.Nil(t, msg.Data())
			}).
			Return(nil).Once()

		p := publisher.New(nc)
		err := p.Publish(context.Background(), "test.subject", nil)
		require.NoError(t, err)
	})
}

func BenchmarkPublisher(b *testing.B) {
	structPayload := &testPayload{Seq: 1, Str: "parson had a dog"}
	protoPayload := wrapperspb.String("parson had a dog")

	benches := []struct {
		name    string
		payload any
		opts    []publisher.PublishOption
	}{
		{"json", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.JSON)}},
		{"json+zstd", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.JSON), publisher.WithContentEncoding(codec.Zstd)}},
		{"json+s2", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.JSON), publisher.WithContentEncoding(codec.S2)}},

		{"yaml", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.YAML)}},

		{"msgpack", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.Msgpack)}},
		{"msgpack+zstd", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.Msgpack), publisher.WithContentEncoding(codec.Zstd)}},
		{"msgpack+s2", structPayload, []publisher.PublishOption{publisher.WithContentType(codec.Msgpack), publisher.WithContentEncoding(codec.S2)}},

		{"protojson", protoPayload, []publisher.PublishOption{publisher.WithContentType(codec.Protojson)}},
		{"prototext", protoPayload, []publisher.PublishOption{publisher.WithContentType(codec.Prototext)}},
		{"protobin", protoPayload, []publisher.PublishOption{publisher.WithContentType(codec.Protobin)}},
		{"protobin+zstd", protoPayload, []publisher.PublishOption{publisher.WithContentType(codec.Protobin), publisher.WithContentEncoding(codec.Zstd)}},
		{"protobin+s2", protoPayload, []publisher.PublishOption{publisher.WithContentType(codec.Protobin), publisher.WithContentEncoding(codec.S2)}},
	}

	ns := xtestutil.Server(b)
	nc := xtestutil.Conn(b, ns)
	p := publisher.New(nc)

	for _, bb := range benches {
		b.Run(bb.name, func(b *testing.B) {
			subj := fmt.Sprintf("bench.publisher.%s", bb.name)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				err := p.Publish(b.Context(), subj, bb.payload, bb.opts...)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
