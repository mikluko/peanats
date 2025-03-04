package peabucket

import (
	"context"
	"net/textproto"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPutUpdateEntryImpl struct {
	key string
	hdr textproto.MIMEHeader
	mod *testModel
	rev uint64
}

func (t *testPutUpdateEntryImpl) Key() string {
	return t.key
}

func (t *testPutUpdateEntryImpl) Header() textproto.MIMEHeader {
	return t.hdr
}

func (t *testPutUpdateEntryImpl) Value() *testModel {
	return t.mod
}

func (t *testPutUpdateEntryImpl) Revision() uint64 {
	return t.rev
}

var (
	_ PutEntry[testModel]    = (*testPutUpdateEntryImpl)(nil)
	_ UpdateEntry[testModel] = (*testPutUpdateEntryImpl)(nil)
)

func TestBucket_New(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		nb := newBucketMock(t)
		b := New[testModel](nb)
		require.NotNil(t, b)
	})
}

func TestBucket_Get(t *testing.T) {
	const value = "----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"balooney"}`
	t.Run("simple", func(t *testing.T) {
		key := "parson.had.a.dog"

		ne := newEntryMock(t)
		ne.OnKey().Return(key).Once()
		ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
		ne.OnValue().TypedReturns([]byte(value)).Once()

		nb := newBucketMock(t)
		nb.OnGet(key).TypedReturns(ne, nil)

		b := New[testModel](nb)
		v, err := b.Get(context.TODO(), key)
		require.NoError(t, err)
		require.NotNil(t, v)

		assert.EqualExportedValues(t, testModel{Name: "balooney"}, *v.Value())
	})
	t.Run("with prefix", func(t *testing.T) {
		prefix := "parson"
		key := "had.a.dog"

		ne := newEntryMock(t)
		ne.OnKey().Return(prefix + "." + key).Once()
		ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
		ne.OnValue().TypedReturns([]byte(value)).Once()

		nb := newBucketMock(t)
		nb.OnGet(prefix+"."+key).TypedReturns(ne, nil)

		b := New[testModel](nb, WithKeyPrefix(prefix))
		v, err := b.Get(context.TODO(), key)
		require.NoError(t, err)
		require.NotNil(t, v)

		assert.Equal(t, testModel{Name: "balooney"}, *v.Value())
	})
}

func TestBucket_GetRevision(t *testing.T) {
	const value = "----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"balooney"}`
	key := "parson.had.a.dog"

	ne := newEntryMock(t)
	ne.OnKey().Return(key).Once()
	ne.OnOperation().TypedReturns(jetstream.KeyValuePut).Once()
	ne.OnValue().TypedReturns([]byte(value)).Once()

	nb := newBucketMock(t)
	nb.OnGetRevision(key, uint64(1)).TypedReturns(ne, nil)

	b := New[testModel](nb)
	v, err := b.GetRevision(context.TODO(), key, 1)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, testModel{Name: "balooney"}, *v.Value())
}

func TestBucket_Put(t *testing.T) {
	const expect = "----\r\nContent-Type: application/json\r\nX-Breed: shavka\r\n\r\n" + `{"name":"balooney"}`
	e := testPutUpdateEntryImpl{
		key: "parson.had.a.dog",
		hdr: textproto.MIMEHeader{"X-Breed": []string{"shavka"}},
		mod: &testModel{Name: "balooney"},
	}

	rb := newBucketMock(t)
	rb.OnPut(e.key, []byte(expect)).TypedReturns(uint64(1), nil)

	b := New[testModel](rb)
	rev, err := b.Put(context.TODO(), &e)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), rev)
}

func TestBucket_Update(t *testing.T) {
	e := testPutUpdateEntryImpl{
		key: "parson.had.a.dog",
		hdr: textproto.MIMEHeader{"X-Breed": []string{"shavka"}},
		mod: &testModel{Name: "balooney"},
		rev: 1,
	}

	rb := newBucketMock(t)
	const expect = "----\r\nContent-Type: application/json\r\nX-Breed: shavka\r\n\r\n" + `{"name":"balooney"}`
	rb.OnUpdate(e.key, []byte(expect), 1).TypedReturns(uint64(2), nil)

	b := New[testModel](rb)
	rev, err := b.Update(context.TODO(), &e)

	require.NoError(t, err)
	assert.Equal(t, uint64(2), rev)
}

func TestBucket_Delete(t *testing.T) {
	key := "uid"

	rb := newBucketMock(t)
	rb.OnDelete(key).Return(nil)

	state := New[testModel](rb)
	err := state.Delete(context.TODO(), key)

	require.NoError(t, err)
}
