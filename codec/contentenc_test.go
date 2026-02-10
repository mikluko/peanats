package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type encodingTestModel struct {
	Name  string `json:"name" yaml:"name" msgpack:"name"`
	Value int    `json:"value" yaml:"value" msgpack:"value"`
}

func TestCompressorRoundTrip(t *testing.T) {
	input := []byte(`{"name":"test","value":42}`)
	tests := []struct {
		name     string
		encoding ContentEncoding
	}{
		{"zstd", Zstd},
		{"s2", S2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp, err := ForContentEncoding(tt.encoding)
			require.NoError(t, err)

			compressed, err := comp.Compress(input)
			require.NoError(t, err)
			assert.NotEqual(t, input, compressed)

			decompressed, err := comp.Decompress(compressed)
			require.NoError(t, err)
			assert.Equal(t, input, decompressed)
		})
	}
}

func TestForContentEncoding_Unknown(t *testing.T) {
	_, err := ForContentEncoding(ContentEncoding(999))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnsupportedEncoding)
}

func TestEncodingFromHeader(t *testing.T) {
	tests := []struct {
		name   string
		header Header
		want   ContentEncoding
	}{
		{
			name:   "nil header",
			header: nil,
			want:   0,
		},
		{
			name:   "empty header",
			header: Header{},
			want:   0,
		},
		{
			name:   "no encoding",
			header: Header{HeaderContentType: []string{JSON.String()}},
			want:   0,
		},
		{
			name:   "zstd",
			header: Header{HeaderContentEncoding: []string{"zstd"}},
			want:   Zstd,
		},
		{
			name:   "s2",
			header: Header{HeaderContentEncoding: []string{"s2"}},
			want:   S2,
		},
		{
			name:   "unknown encoding",
			header: Header{HeaderContentEncoding: []string{"gzip"}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, EncodingFromHeader(tt.header))
		})
	}
}

func TestSetContentEncoding(t *testing.T) {
	h := Header{}
	SetContentEncoding(h, Zstd)
	assert.Equal(t, "zstd", h.Get(HeaderContentEncoding))

	SetContentEncoding(h, S2)
	assert.Equal(t, "s2", h.Get(HeaderContentEncoding))
}

func TestMarshalHeaderWithEncoding(t *testing.T) {
	model := encodingTestModel{Name: "test", Value: 42}
	encodings := []struct {
		name     string
		encoding ContentEncoding
	}{
		{"zstd", Zstd},
		{"s2", S2},
	}
	for _, enc := range encodings {
		t.Run(enc.name, func(t *testing.T) {
			header := Header{
				HeaderContentType:     []string{JSON.String()},
				HeaderContentEncoding: []string{enc.encoding.String()},
			}

			data, err := MarshalHeader(model, header)
			require.NoError(t, err)

			// Compressed data should differ from plain JSON
			plainHeader := Header{HeaderContentType: []string{JSON.String()}}
			plainData, err := MarshalHeader(model, plainHeader)
			require.NoError(t, err)
			assert.NotEqual(t, plainData, data)

			// Round-trip: unmarshal should decompress and decode
			var result encodingTestModel
			err = UnmarshalHeader(data, &result, header)
			require.NoError(t, err)
			assert.Equal(t, model, result)
		})
	}
}

func TestMarshalHeaderWithEncoding_NilValue(t *testing.T) {
	header := Header{
		HeaderContentType:     []string{JSON.String()},
		HeaderContentEncoding: []string{Zstd.String()},
	}
	data, err := MarshalHeader(nil, header)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestUnmarshalHeaderWithEncoding_NilData(t *testing.T) {
	header := Header{
		HeaderContentType:     []string{JSON.String()},
		HeaderContentEncoding: []string{Zstd.String()},
	}
	var result encodingTestModel
	err := UnmarshalHeader(nil, &result, header)
	require.NoError(t, err)
}

func TestContentEncodingString(t *testing.T) {
	assert.Equal(t, "zstd", Zstd.String())
	assert.Equal(t, "s2", S2.String())
	assert.Panics(t, func() { ContentEncoding(999).String() })
}
