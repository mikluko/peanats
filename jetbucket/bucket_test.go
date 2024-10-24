package jetbucket

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucket_New(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		nb := newBucketMock(t)
		b := NewBucket[TestModel](nb)
		require.NotNil(t, b)
	})
	t.Run("panics", func(t *testing.T) {
		nb := newBucketMock(t)
		require.Panics(t, func() {
			_ = NewBucket[struct{}](nb)
		})
	})
}

func TestBucket_Get(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		key := "parson-had-a-dog"

		ne := newEntryMock(t)
		ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
		ne.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

		nb := newBucketMock(t)
		nb.OnGet(key).TypedReturns(ne, nil)

		b := NewBucket[TestModel](nb)
		ent, val, err := b.Get(context.TODO(), key)
		require.NoError(t, err)
		require.NotNil(t, ent)
		require.NotNil(t, val)

		assert.Equal(t, "balooney", val.Name)
	})
	t.Run("prefixed", func(t *testing.T) {
		prefix := "parson"
		key := "had.a.dog"

		ne := newEntryMock(t)
		ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
		ne.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

		nb := newBucketMock(t)
		nb.OnGet(prefix+"."+key).TypedReturns(ne, nil)

		b := NewBucket[TestModel](nb, WithKeyPrefix(prefix))
		ent, val, err := b.Get(context.TODO(), key)
		require.NoError(t, err)
		require.NotNil(t, ent)
		require.NotNil(t, val)

		assert.Equal(t, "balooney", val.Name)
	})
}

func TestBucket_Put(t *testing.T) {
	key := "parson-had-a-dog"
	mod := TestModel{Name: "balooney"}

	rb := newBucketMock(t)
	rb.OnPut(key, []byte(`{"name":"balooney"}`)).TypedReturns(uint64(1), nil)

	b := NewBucket[TestModel](rb)
	rev, err := b.Put(context.TODO(), key, &mod)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), rev)
}

func TestBucket_Delete(t *testing.T) {
	key := "uid"

	rb := newBucketMock(t)
	rb.OnDelete(key).Return(nil)

	state := NewBucket[TestModel](rb)
	err := state.Delete(context.TODO(), key)

	require.NoError(t, err)
}
