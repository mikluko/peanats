package jetbucket

import (
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	up0 := newEntryMock(t)
	up0.OnKey().Return("parson.had.a.dog").Once()
	up0.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
	up0.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

	up1 := newEntryMock(t)
	up1.OnKey().Return("parson.had.a.cat").Once()
	up1.OnOperation().TypedReturns(jetstream.KeyValueDelete).Once()

	ch := make(chan jetstream.KeyValueEntry, 3)
	defer close(ch)

	ch <- up0
	ch <- nil
	ch <- up1

	nw := newWatcherMock(t)
	nw.OnUpdates().TypedReturns(ch)

	w := NewWatcher[TestModel](nw, WatcherPrefix("parson"))

	e, err := w.Next()
	require.NoError(t, err)
	assert.NotNil(t, e.Value())
	assert.Equal(t, "had.a.dog", e.Key())
	assert.Equal(t, TestModel{Name: "balooney"}, *e.Value())

	e, err = w.Next()
	require.ErrorIs(t, err, ErrInitialValuesOver)
	assert.Nil(t, e)

	e, err = w.Next()
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Equal(t, "had.a.cat", e.Key())
	assert.Nil(t, e.Value())
}
