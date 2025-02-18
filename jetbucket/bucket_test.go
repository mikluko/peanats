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
		key := "parson.had.a.dog"

		ne := newEntryMock(t)
		ne.OnKey().Return(key).Once()
		ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
		ne.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

		nb := newBucketMock(t)
		nb.OnGet(key).TypedReturns(ne, nil)

		b := NewBucket[TestModel](nb)
		v, err := b.Get(context.TODO(), key)
		require.NoError(t, err)
		require.NotNil(t, v)

		assert.Equal(t, TestModel{Name: "balooney"}, *v.Value())
	})
	t.Run("with prefix", func(t *testing.T) {
		prefix := "parson"
		key := "had.a.dog"

		ne := newEntryMock(t)
		ne.OnKey().Return(prefix + "." + key).Once()
		ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
		ne.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

		nb := newBucketMock(t)
		nb.OnGet(prefix+"."+key).TypedReturns(ne, nil)

		b := NewBucket[TestModel](nb, WithKeyPrefix(prefix))
		v, err := b.Get(context.TODO(), key)
		require.NoError(t, err)
		require.NotNil(t, v)

		assert.Equal(t, TestModel{Name: "balooney"}, *v.Value())
	})
}

func TestBucket_GetRevision(t *testing.T) {
	key := "parson.had.a.dog"

	ne := newEntryMock(t)
	ne.OnKey().Return(key).Once()
	ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
	ne.OnValue().TypedReturns([]byte(`{"name":"balooney"}`)).Once()

	nb := newBucketMock(t)
	nb.OnGetRevision(key, uint64(1)).TypedReturns(ne, nil)

	b := NewBucket[TestModel](nb)
	v, err := b.GetRevision(context.TODO(), key, 1)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, TestModel{Name: "balooney"}, *v.Value())
}

func TestBucket_Put(t *testing.T) {
	key := "parson.had.a.dog"
	mod := TestModel{Name: "balooney"}

	rb := newBucketMock(t)
	rb.OnPut(key, []byte(`{"name":"balooney"}`)).TypedReturns(uint64(1), nil)

	b := NewBucket[TestModel](rb)
	rev, err := b.Put(context.TODO(), key, &mod)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), rev)
}

func TestBucket_Update(t *testing.T) {
	key := "parson.had.a.dog"
	mod := TestModel{Name: "balooney"}

	rb := newBucketMock(t)
	rb.OnUpdate(key, []byte(`{"name":"balooney"}`), 1).TypedReturns(uint64(2), nil)

	b := NewBucket[TestModel](rb)
	rev, err := b.Update(context.TODO(), key, &mod, 1)

	require.NoError(t, err)
	assert.Equal(t, uint64(2), rev)
}

func TestBucket_Delete(t *testing.T) {
	key := "uid"

	rb := newBucketMock(t)
	rb.OnDelete(key).Return(nil)

	state := NewBucket[TestModel](rb)
	err := state.Delete(context.TODO(), key)

	require.NoError(t, err)
}
