package peanats

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
		name   string
		header Header
		model  any
		want   string
	}{
		{
			name:  "default",
			model: model{"bar"},
			want:  `{"foo":"bar"}`,
		},
		{
			name:   "json",
			header: Header{HeaderContentType: []string{ContentTypeJson.String()}},
			model:  model{"bar"},
			want:   `{"foo":"bar"}`,
		},
		{
			name:   "yaml",
			header: Header{HeaderContentType: []string{ContentTypeYaml.String()}},
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

func TestContentTypeHeaderCopy(t *testing.T) {
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
			want: ContentTypeJson.String(),
		},
		{
			name: "non-default",
			dst:  Header{HeaderContentType: []string{ContentTypeJson.String()}},
			src:  Header{HeaderContentType: []string{ContentTypeYaml.String()}},
			want: ContentTypeYaml.String(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ContentTypeHeaderCopy(tt.dst, tt.src)
			assert.Equal(t, tt.want, tt.dst.Get(HeaderContentType))
		})
	}
}
