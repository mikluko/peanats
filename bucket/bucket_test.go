package bucket_test

import (
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/bucket"
	"github.com/mikluko/peanats/internal/xmock/jetstreammock"
)

type testModel struct {
	Name string `json:"name"`
}

type testPutUpdateEntryImpl struct {
	key string
	hdr peanats.Header
	mod *testModel
	rev uint64
}

func (t *testPutUpdateEntryImpl) Key() string {
	return t.key
}

func (t *testPutUpdateEntryImpl) Header() peanats.Header {
	return t.hdr
}

func (t *testPutUpdateEntryImpl) Value() *testModel {
	return t.mod
}

func (t *testPutUpdateEntryImpl) Revision() uint64 {
	return t.rev
}

var (
	_ bucket.PutBucketEntry[testModel]    = (*testPutUpdateEntryImpl)(nil)
	_ bucket.UpdateBucketEntry[testModel] = (*testPutUpdateEntryImpl)(nil)
)

func TestBucket_New(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		nb := jetstreammock.NewKeyValue(t)
		b := bucket.NewBucket[testModel](nb)
		require.NotNil(t, b)
	})
}

func TestBucket_Get(t *testing.T) {
	const value = "----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"balooney"}`
	t.Run("simple", func(t *testing.T) {
		key := "parson.had.a.dog"

		ne := jetstreammock.NewKeyValueEntry(t)
		ne.EXPECT().Key().Return(key).Once()
		ne.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
		ne.EXPECT().Value().Return([]byte(value)).Once()

		nb := jetstreammock.NewKeyValue(t)
		nb.EXPECT().Get(mock.Anything, key).Return(ne, nil)

		b := bucket.NewBucket[testModel](nb)
		v, err := b.Get(t.Context(), key)
		require.NoError(t, err)
		require.NotNil(t, v)
		assert.Equal(t, key, v.Key())

		assert.EqualExportedValues(t, testModel{Name: "balooney"}, *v.Value())
	})
	t.Run("with prefix", func(t *testing.T) {
		prefix := "parson"
		key := "had.a.dog"

		ne := jetstreammock.NewKeyValueEntry(t)
		ne.EXPECT().Key().Return(prefix + "." + key).Once()
		ne.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
		ne.EXPECT().Value().Return([]byte(value)).Once()

		nb := jetstreammock.NewKeyValue(t)
		nb.EXPECT().Get(mock.Anything, prefix+"."+key).Return(ne, nil)

		b := bucket.NewBucket[testModel](nb, bucket.BucketKeyPrefix(prefix))
		v, err := b.Get(t.Context(), key)
		require.NoError(t, err)
		require.NotNil(t, v)
		assert.Equal(t, key, v.Key())

		assert.Equal(t, testModel{Name: "balooney"}, *v.Value())
	})
}

func TestBucket_GetRevision(t *testing.T) {
	const value = "----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"balooney"}`
	key := "parson.had.a.dog"

	ne := jetstreammock.NewKeyValueEntry(t)
	ne.EXPECT().Key().Return(key).Once()
	ne.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
	ne.EXPECT().Value().Return([]byte(value)).Once()

	nb := jetstreammock.NewKeyValue(t)
	nb.EXPECT().GetRevision(mock.Anything, key, uint64(1)).Return(ne, nil)

	b := bucket.NewBucket[testModel](nb)
	v, err := b.GetRevision(t.Context(), key, 1)
	require.NoError(t, err)
	require.NotNil(t, v)
	assert.Equal(t, key, v.Key())

	assert.Equal(t, testModel{Name: "balooney"}, *v.Value())
}

func TestBucket_Put(t *testing.T) {
	const expect = "----\r\nContent-Type: application/json\r\nX-Breed: shavka\r\n\r\n" + `{"name":"balooney"}`
	e := testPutUpdateEntryImpl{
		key: "parson.had.a.dog",
		hdr: peanats.Header{"X-Breed": []string{"shavka"}},
		mod: &testModel{Name: "balooney"},
	}

	nb := jetstreammock.NewKeyValue(t)
	nb.EXPECT().Put(mock.Anything, e.key, []byte(expect)).Return(uint64(1), nil)

	b := bucket.NewBucket[testModel](nb)
	rev, err := b.Put(t.Context(), &e)

	require.NoError(t, err)
	assert.Equal(t, uint64(1), rev)
}

func TestBucket_Update(t *testing.T) {
	e := testPutUpdateEntryImpl{
		key: "parson.had.a.dog",
		hdr: peanats.Header{"X-Breed": []string{"shavka"}},
		mod: &testModel{Name: "balooney"},
		rev: 1,
	}

	nb := jetstreammock.NewKeyValue(t)
	const expect = "----\r\nContent-Type: application/json\r\nX-Breed: shavka\r\n\r\n" + `{"name":"balooney"}`
	nb.EXPECT().Update(mock.Anything, e.key, []byte(expect), uint64(1)).Return(uint64(2), nil)

	b := bucket.NewBucket[testModel](nb)
	rev, err := b.Update(t.Context(), &e)

	require.NoError(t, err)
	assert.Equal(t, uint64(2), rev)
}

func TestBucket_Delete(t *testing.T) {
	key := "uid"

	nb := jetstreammock.NewKeyValue(t)
	nb.EXPECT().Delete(mock.Anything, key).Return(nil)

	state := bucket.NewBucket[testModel](nb)
	err := state.Delete(t.Context(), key)

	require.NoError(t, err)
}
