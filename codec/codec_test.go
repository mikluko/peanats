package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalHeader(t *testing.T) {
	type model struct {
		Foo string `json:"foo"`
	}
	tests := []struct {
		name       string
		header     Header
		model      any
		want       string
		wantHeader Header
	}{
		{
			name:   "default",
			header: Header{},
			model:  model{"bar"},
			want:   `{"foo":"bar"}`,
		},
		{
			name:   "default nil",
			header: nil,
			model:  model{"bar"},
			want:   `{"foo":"bar"}`,
		},
		{
			name:   "json",
			header: Header{HeaderContentType: []string{JSON.String()}},
			model:  model{"bar"},
			want:   `{"foo":"bar"}`,
		},
		{
			name:   "yaml",
			header: Header{HeaderContentType: []string{YAML.String()}},
			model:  model{"bar"},
			want:   "foo: bar\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalHeader(tt.model, tt.header)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestTypeFromHeaderCopy(t *testing.T) {
	tests := []struct {
		name string
		dst  Header
		src  Header
		want string
	}{
		{
			name: "default",
			dst:  Header{},
			src:  Header{},
			want: JSON.String(),
		},
		{
			name: "non-default",
			dst:  Header{HeaderContentType: []string{JSON.String()}},
			src:  Header{HeaderContentType: []string{YAML.String()}},
			want: YAML.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TypeFromHeaderCopy(tt.dst, tt.src)
			assert.Equal(t, tt.want, tt.dst.Get(HeaderContentType))
		})
	}
}
