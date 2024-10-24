package jetbucket

import (
	"encoding/json"
	"errors"
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
	t.Run("ok", func(t *testing.T) {
		m := &TestModel{Name: "test"}
		b, err := marshal(m)
		require.NoError(t, err)
		assert.Equal(t, `{"name":"test"}`, string(b))
	})
	t.Run("error/unimplemented", func(t *testing.T) {
		m := struct{}{}
		b, err := marshal(m)
		require.ErrorIs(t, err, ErrMarshal)
		assert.Nil(t, b)
	})
	t.Run("error/marshal", func(t *testing.T) {
		m := newMarshalerMock(t)
		m.OnMarshal().Return(nil, errors.New("some error")).Once()
		b, err := marshal(m)
		require.ErrorIs(t, err, ErrMarshal)
		assert.Nil(t, b)
	})
}

func TestUnmarshal(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		m := &TestModel{}
		err := unmarshal(m, []byte(`{"name":"test"}`))
		require.NoError(t, err)
		assert.Equal(t, "test", m.Name)
	})
	t.Run("error/unimplemented", func(t *testing.T) {
		m := struct{}{}
		err := unmarshal(m, []byte(`{"name":"test"}`))
		require.ErrorIs(t, err, ErrUnmarshal)
	})
	t.Run("error/marshal", func(t *testing.T) {
		m := newUnmarshalerMock(t)
		m.OnUnmarshal([]byte(`{"name":"test"}`)).Return(errors.New("some error")).Once()
		err := unmarshal(m, []byte(`{"name":"test"}`))
		require.ErrorIs(t, err, ErrUnmarshal)
	})
	t.Run("error/malformed", func(t *testing.T) {
		m := &TestModel{}
		err := unmarshal(m, []byte(`}{yy`))
		require.ErrorIs(t, err, ErrUnmarshal)
	})
}
