package bucket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikluko/peanats"
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
