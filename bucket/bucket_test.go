package bucket_test

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/bucket"
	"github.com/mikluko/peanats/codec"
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
	_ bucket.PutEntry[testModel]    = (*testPutUpdateEntryImpl)(nil)
	_ bucket.UpdateEntry[testModel] = (*testPutUpdateEntryImpl)(nil)
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

func TestBucket_GetLatestRevision(t *testing.T) {
	const value = "----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"balooney"}`
	key := "parson.had.a.dog"

	t.Run("key not found", func(t *testing.T) {
		// Create a mock watcher with updates channel
		nw := jetstreammock.NewKeyWatcher(t)
		updates := make(chan jetstream.KeyValueEntry, 1)
		updates <- nil // Simulate no revisions
		nw.EXPECT().Updates().Return(updates).Once()
		nw.EXPECT().Stop().Once().Return(nil)

		// Set up the KeyValue mock to return our mock watcher
		nb := jetstreammock.NewKeyValue(t)
		nb.EXPECT().Watch(mock.Anything, key).Return(nw, nil).Once()

		// Create the bucket and test GetLatestRevision
		b := bucket.NewBucket[testModel](nb)
		v, err := b.GetLatestRevision(t.Context(), key)

		// Verify results
		require.ErrorIs(t, err, jetstream.ErrKeyNotFound)
		require.Nil(t, v)
	})

	t.Run("key exists", func(t *testing.T) {
		// Create a mock KeyValueEntry
		ne := jetstreammock.NewKeyValueEntry(t)
		ne.EXPECT().Key().Return(key).Once()
		ne.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
		ne.EXPECT().Value().Return([]byte(value)).Once()
		ne.EXPECT().Revision().Return(1).Once()

		// Create a mock watcher with updates channel
		nw := jetstreammock.NewKeyWatcher(t)

		// Set up the updates channel to return our mock entry
		updates := make(chan jetstream.KeyValueEntry, 1)
		updates <- ne
		nw.EXPECT().Updates().Return(updates).Once()
		nw.EXPECT().Stop().Once().Return(nil)

		// Set up the KeyValue mock to return our mock watcher
		nb := jetstreammock.NewKeyValue(t)
		nb.EXPECT().Watch(mock.Anything, key).Return(nw, nil).Once()

		// Create the bucket and test GetLatestRevision
		b := bucket.NewBucket[testModel](nb)
		v, err := b.GetLatestRevision(t.Context(), key)

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, v)
		assert.Equal(t, key, v.Key())
		assert.Equal(t, testModel{Name: "balooney"}, *v.Value())
		assert.Equal(t, uint64(1), v.Revision())
	})
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

func TestBucket_History(t *testing.T) {
	t.Run("get history entries", func(t *testing.T) {
		// Define prefix and keys
		prefix := "parson"
		unprefixedKey := "had.a.dog"
		prefixedKey := prefix + "." + unprefixedKey

		// Create multiple mock entries representing history revisions
		ne1 := jetstreammock.NewKeyValueEntry(t)
		ne1.EXPECT().Key().Return(prefixedKey).Once()
		ne1.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
		ne1.EXPECT().Value().Return([]byte("----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"spot"}`)).Once()
		ne1.EXPECT().Revision().Return(uint64(1)).Once()

		ne2 := jetstreammock.NewKeyValueEntry(t)
		ne2.EXPECT().Key().Return(prefixedKey).Once()
		ne2.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
		ne2.EXPECT().Value().Return([]byte("----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"rover"}`)).Once()
		ne2.EXPECT().Revision().Return(uint64(2)).Once()

		ne3 := jetstreammock.NewKeyValueEntry(t)
		ne3.EXPECT().Key().Return(prefixedKey).Once()
		ne3.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
		ne3.EXPECT().Value().Return([]byte("----\r\nContent-Type: application/json\r\n\r\n" + `{"name":"balooney"}`)).Once()
		ne3.EXPECT().Revision().Return(uint64(3)).Once()

		// Create a slice of history entries
		historyEntries := []jetstream.KeyValueEntry{ne1, ne2, ne3}

		// Set up the KeyValue mock to return our history entries
		nb := jetstreammock.NewKeyValue(t)
		nb.EXPECT().History(mock.Anything, prefixedKey, mock.Anything).Return(historyEntries, nil).Once()

		// Create the bucket with prefix and test History
		b := bucket.NewBucket[testModel](nb, bucket.BucketKeyPrefix(prefix))
		entries, err := b.History(t.Context(), unprefixedKey)

		// Verify results
		require.NoError(t, err)
		require.Len(t, entries, 3)

		// Verify the entries are in the correct order (oldest to newest)
		// and that keys are properly deprefixed
		assert.Equal(t, unprefixedKey, entries[0].Key())
		assert.Equal(t, "spot", entries[0].Value().Name)
		assert.Equal(t, uint64(1), entries[0].Revision())

		assert.Equal(t, unprefixedKey, entries[1].Key())
		assert.Equal(t, "rover", entries[1].Value().Name)
		assert.Equal(t, uint64(2), entries[1].Revision())

		assert.Equal(t, unprefixedKey, entries[2].Key())
		assert.Equal(t, "balooney", entries[2].Value().Name)
		assert.Equal(t, uint64(3), entries[2].Revision())
	})

	t.Run("empty history", func(t *testing.T) {
		key := "no-history-key"

		// Set up the KeyValue mock to return empty history
		nb := jetstreammock.NewKeyValue(t)
		nb.EXPECT().History(mock.Anything, key, mock.Anything).Return([]jetstream.KeyValueEntry{}, nil).Once()

		// Create the bucket and test History
		b := bucket.NewBucket[testModel](nb)
		entries, err := b.History(t.Context(), key)

		// Verify results
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

func TestBucket_PutGetWithEncoding(t *testing.T) {
	encodings := []struct {
		name     string
		encoding codec.ContentEncoding
	}{
		{"zstd", codec.Zstd},
		{"s2", codec.S2},
	}
	for _, enc := range encodings {
		t.Run(enc.name, func(t *testing.T) {
			key := "parson.had.a.dog"
			e := testPutUpdateEntryImpl{
				key: key,
				hdr: peanats.Header{},
				mod: &testModel{Name: "balooney"},
			}

			// Get uncompressed reference for comparison
			ePlain := testPutUpdateEntryImpl{
				key: key,
				hdr: peanats.Header{},
				mod: &testModel{Name: "balooney"},
			}
			var plainCaptured []byte
			nbPlain := jetstreammock.NewKeyValue(t)
			nbPlain.EXPECT().
				Put(mock.Anything, key, mock.Anything).
				Run(func(_ context.Context, _ string, data []byte) {
					plainCaptured = make([]byte, len(data))
					copy(plainCaptured, data)
				}).
				Return(uint64(1), nil).Once()
			bPlain := bucket.NewBucket[testModel](nbPlain)
			_, err := bPlain.Put(t.Context(), &ePlain)
			require.NoError(t, err)

			// Capture the bytes written by Put with compression
			var captured []byte
			nb := jetstreammock.NewKeyValue(t)
			nb.EXPECT().
				Put(mock.Anything, key, mock.Anything).
				Run(func(_ context.Context, _ string, data []byte) {
					captured = make([]byte, len(data))
					copy(captured, data)
				}).
				Return(uint64(1), nil).Once()

			b := bucket.NewBucket[testModel](nb, bucket.BucketContentEncoding(enc.encoding))
			rev, err := b.Put(t.Context(), &e)
			require.NoError(t, err)
			assert.Equal(t, uint64(1), rev)

			// Compressed output should differ from plain
			assert.NotEqual(t, plainCaptured, captured)
			assert.Contains(t, string(captured), "Content-Encoding: "+enc.encoding.String())

			// Round-trip: Get should decompress and decode
			ne := jetstreammock.NewKeyValueEntry(t)
			ne.EXPECT().Key().Return(key).Once()
			ne.EXPECT().Operation().Return(jetstream.KeyValuePut).Once()
			ne.EXPECT().Value().Return(captured).Once()

			nb2 := jetstreammock.NewKeyValue(t)
			nb2.EXPECT().Get(mock.Anything, key).Return(ne, nil).Once()

			b2 := bucket.NewBucket[testModel](nb2)
			v, err := b2.Get(t.Context(), key)
			require.NoError(t, err)
			assert.Equal(t, "balooney", v.Value().Name)
		})
	}
}

func TestBucket_UpdateWithEncoding(t *testing.T) {
	key := "parson.had.a.dog"

	// Get uncompressed reference
	ePlain := testPutUpdateEntryImpl{
		key: key, hdr: peanats.Header{}, mod: &testModel{Name: "balooney"}, rev: 1,
	}
	var plainCaptured []byte
	nbPlain := jetstreammock.NewKeyValue(t)
	nbPlain.EXPECT().
		Update(mock.Anything, key, mock.Anything, uint64(1)).
		Run(func(_ context.Context, _ string, data []byte, _ uint64) {
			plainCaptured = make([]byte, len(data))
			copy(plainCaptured, data)
		}).
		Return(uint64(2), nil).Once()
	bPlain := bucket.NewBucket[testModel](nbPlain)
	_, err := bPlain.Update(t.Context(), &ePlain)
	require.NoError(t, err)

	// Update with compression
	e := testPutUpdateEntryImpl{
		key: key, hdr: peanats.Header{}, mod: &testModel{Name: "balooney"}, rev: 1,
	}
	var captured []byte
	nb := jetstreammock.NewKeyValue(t)
	nb.EXPECT().
		Update(mock.Anything, key, mock.Anything, uint64(1)).
		Run(func(_ context.Context, _ string, data []byte, _ uint64) {
			captured = make([]byte, len(data))
			copy(captured, data)
		}).
		Return(uint64(2), nil).Once()

	b := bucket.NewBucket[testModel](nb, bucket.BucketContentEncoding(codec.Zstd))
	rev, err := b.Update(t.Context(), &e)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), rev)

	// Compressed output should differ from plain
	assert.NotEqual(t, plainCaptured, captured)
	assert.Contains(t, string(captured), "Content-Encoding: zstd")
}
