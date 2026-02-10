package bucket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
	"github.com/mikluko/peanats/codec"
)

type testModel struct {
	Name string `json:"name"`
}

func TestEncode(t *testing.T) {
	m := testModel{Name: "balooney"}
	b, err := encodeBucketEntryHeader(peanats.Header{"X-Breed": []string{"shavka"}}, &m)
	require.NoError(t, err)
	assert.Contains(t, string(b), `Content-Type: application/json`)
	assert.Contains(t, string(b), `X-Breed: shavka`)
	assert.Contains(t, string(b), `{"name":"balooney"}`)
}

func TestDecode(t *testing.T) {
	const data = "----\nContent-Type: application/json\nX-Breed: shavka\n\n" + `{"name":"balooney"}`
	h, v, err := decodeBucketEntryHeader[testModel]([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "shavka", h.Get("x-breed"))
	assert.Equal(t, "balooney", v.Name)
}

func TestEncodeDecodeWithEncoding(t *testing.T) {
	encodings := []struct {
		name     string
		encoding codec.ContentEncoding
	}{
		{"zstd", codec.Zstd},
		{"s2", codec.S2},
	}
	for _, enc := range encodings {
		t.Run(enc.name, func(t *testing.T) {
			m := testModel{Name: "balooney"}

			// Encode without compression for comparison
			plainHeader := peanats.Header{"X-Breed": []string{"shavka"}}
			plainData, err := encodeBucketEntryHeader(plainHeader, &m)
			require.NoError(t, err)

			// Encode with compression
			compHeader := peanats.Header{"X-Breed": []string{"shavka"}}
			codec.SetContentEncoding(compHeader, enc.encoding)
			compData, err := encodeBucketEntryHeader(compHeader, &m)
			require.NoError(t, err)

			// Compressed output should differ from plain
			assert.NotEqual(t, plainData, compData)
			// Headers should still be readable in the multipart envelope
			assert.Contains(t, string(compData), "Content-Encoding: "+enc.encoding.String())
			assert.Contains(t, string(compData), "Content-Type: application/json")

			// Round-trip decode
			dh, v, err := decodeBucketEntryHeader[testModel](compData)
			require.NoError(t, err)
			assert.Equal(t, "balooney", v.Name)
			assert.Equal(t, "shavka", dh.Get("x-breed"))
			assert.Equal(t, enc.encoding.String(), dh.Get(codec.HeaderContentEncoding))
		})
	}
}
