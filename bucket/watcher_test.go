package bucket_test

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats/bucket"
	"github.com/mikluko/peanats/internal/xmock/bucketmock"
	"github.com/mikluko/peanats/internal/xmock/jetstreammock"
)

func TestWatcher(t *testing.T) {
	up0 := jetstreammock.NewKeyValueEntry(t)
	up0.EXPECT().Key().Return("parson.had.a.dog").Once()
	up0.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
	up0.EXPECT().Value().Return([]byte("----\r\nContent-Type: application/json\r\nX-Breed: shavka\r\n\r\n" + `{"name":"balooney"}`)).Once()

	up1 := jetstreammock.NewKeyValueEntry(t)
	up1.EXPECT().Key().Return("parson.had.a.cat").Once()
	up1.EXPECT().Operation().Return(jetstream.KeyValueDelete).Once()

	ch := make(chan jetstream.KeyValueEntry, 3)
	defer close(ch)

	ch <- up0
	ch <- nil
	ch <- up1

	nw := jetstreammock.NewKeyWatcher(t)
	nw.EXPECT().Updates().Return(ch)

	w := bucket.NewWatcher[testModel](nw, bucket.WatcherPrefix("parson"))

	e, err := w.Next()
	require.NoError(t, err)
	assert.NotNil(t, e.Value())
	assert.Equal(t, "had.a.dog", e.Key())
	assert.Equal(t, testModel{Name: "balooney"}, *e.Value())

	e, err = w.Next()
	require.ErrorIs(t, err, bucket.ErrInitialValuesOver)
	assert.Nil(t, e)

	e, err = w.Next()
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, "had.a.cat", e.Key())
	assert.Nil(t, e.Value())
}

func TestWatch(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		m := bucketmock.NewEntry[testModel](t)
		m.EXPECT().Key().Return("parson.had.a.dog")
		m.EXPECT().Value().Return(&testModel{Name: "balooney"})

		w := bucketmock.NewWatcher[testModel](t)

		w.EXPECT().Next().Return(m, nil).Once()
		w.EXPECT().Next().Return(nil, bucket.ErrInitialValuesOver).Once()
		w.EXPECT().Next().Return(m, nil).Once()
		w.EXPECT().Next().Return(nil, io.EOF).Once()

		wg := sync.WaitGroup{}
		wg.Add(2)

		h := bucket.EntryHandlerFunc[testModel](func(ctx context.Context, b bucket.Entry[testModel]) error {
			assert.Equal(t, "parson.had.a.dog", b.Key())
			assert.Equal(t, &testModel{Name: "balooney"}, b.Value())
			wg.Done()
			return nil
		})

		err := bucket.Watch(t.Context(), w, h)
		require.NoError(t, err)

		wg.Wait()
	})
}
