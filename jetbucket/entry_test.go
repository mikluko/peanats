package jetbucket

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	Name string `json:"name"`
}

func (m *TestModel) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m *TestModel) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

func TestMarshal(t *testing.T) {
	m := &TestModel{Name: "test"}
	b, err := marshal(m)
	require.NoError(t, err)
	assert.Equal(t, `{"name":"test"}`, string(b))
}

func TestUnmarshal(t *testing.T) {
	m := &TestModel{}
	err := unmarshal(m, []byte(`{"name":"test"}`))
	require.NoError(t, err)
	assert.Equal(t, "test", m.Name)
}
