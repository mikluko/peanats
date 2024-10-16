package jetbucket

import (
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatcher(t *testing.T) {
	up0 := newEntryMock(t)
	up0.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
	up0.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

	up1 := newEntryMock(t)
	up1.OnOperation().TypedReturns(jetstream.KeyValueDelete).Once()

	ch := make(chan jetstream.KeyValueEntry, 3)
	defer close(ch)

	ch <- up0
	ch <- nil
	ch <- up1

	nw := newWatcherMock(t)
	nw.OnUpdates().TypedReturns(ch)

	w := NewWatcher[TestModel](nw)

	e, v, err := w.Next()
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.NotNil(t, v)
	assert.Equal(t, "balooney", v.Name)

	e, v, err = w.Next()
	require.ErrorIs(t, err, ErrInitialValuesOver)
	assert.Nil(t, e)
	assert.Nil(t, v)

	e, v, err = w.Next()
	require.NoError(t, err)
	assert.NotNil(t, e)
	assert.Nil(t, v)
}
