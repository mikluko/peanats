package jetbucket

import (
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testModel struct {
	Name string `json:"name"`
}

func TestEncode(t *testing.T) {
	m := testModel{Name: "balooney"}
	b, err := encode(textproto.MIMEHeader{"X-Breed": []string{"shavka"}}, &m)
	require.NoError(t, err)
	assert.Contains(t, string(b), `Content-Type: application/json`)
	assert.Contains(t, string(b), `X-Breed: shavka`)
	assert.Contains(t, string(b), `{"name":"balooney"}`)
}

func TestDecode(t *testing.T) {
	const data = "----\nContent-Type: application/json\nX-Breed: shavka\n\n" + `{"name":"balooney"}`
	h, v, err := decode[testModel]([]byte(data))
	require.NoError(t, err)
	assert.Equal(t, "shavka", h.Get("x-breed"))
	assert.Equal(t, "balooney", v.Name)
}
