package codec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
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

func TestProtoCodecRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		contentType ContentType
	}{
		{"protojson", Protojson},
		{"prototext", Prototext},
		{"protobin", Protobin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := Header{HeaderContentType: []string{tt.contentType.String()}}
			original := wrapperspb.String("parson had a dog")

			data, err := MarshalHeader(original, header)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			result := &wrapperspb.StringValue{}
			err = UnmarshalHeader(data, result, header)
			require.NoError(t, err)
			assert.Equal(t, original.GetValue(), result.GetValue())
		})
	}
}

func TestProtoCodecNonProtoMessage(t *testing.T) {
	type notProto struct {
		Foo string
	}
	tests := []struct {
		name        string
		contentType ContentType
	}{
		{"protojson", Protojson},
		{"prototext", Prototext},
		{"protobin", Protobin},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/marshal", func(t *testing.T) {
			header := Header{HeaderContentType: []string{tt.contentType.String()}}
			_, err := MarshalHeader(notProto{Foo: "bar"}, header)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "is not a proto.Message")
		})
		t.Run(tt.name+"/unmarshal", func(t *testing.T) {
			header := Header{HeaderContentType: []string{tt.contentType.String()}}
			err := UnmarshalHeader([]byte("data"), &notProto{}, header)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "is not a proto.Message")
		})
	}
}

func BenchmarkMarshalHeader(b *testing.B) {
	type model struct {
		Foo string `json:"foo" yaml:"foo" msgpack:"foo"`
	}
	structPayload := model{Foo: "parson had a dog"}
	protoPayload := wrapperspb.String("parson had a dog")

	codecs := []struct {
		name    string
		ct      ContentType
		payload any
	}{
		{"json", JSON, structPayload},
		{"yaml", YAML, structPayload},
		{"msgpack", Msgpack, structPayload},
		{"protojson", Protojson, protoPayload},
		{"prototext", Prototext, protoPayload},
		{"protobin", Protobin, protoPayload},
	}
	for _, cc := range codecs {
		header := Header{HeaderContentType: []string{cc.ct.String()}}
		b.Run(cc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _ = MarshalHeader(cc.payload, header)
			}
		})
	}
}

func BenchmarkUnmarshalHeader(b *testing.B) {
	type model struct {
		Foo string `json:"foo" yaml:"foo" msgpack:"foo"`
	}
	structPayload := model{Foo: "parson had a dog"}
	protoPayload := wrapperspb.String("parson had a dog")

	codecs := []struct {
		name    string
		ct      ContentType
		payload any
		newDst  func() any
	}{
		{"json", JSON, structPayload, func() any { return &model{} }},
		{"yaml", YAML, structPayload, func() any { return &model{} }},
		{"msgpack", Msgpack, structPayload, func() any { return &model{} }},
		{"protojson", Protojson, protoPayload, func() any { return &wrapperspb.StringValue{} }},
		{"prototext", Prototext, protoPayload, func() any { return &wrapperspb.StringValue{} }},
		{"protobin", Protobin, protoPayload, func() any { return &wrapperspb.StringValue{} }},
	}
	for _, cc := range codecs {
		header := Header{HeaderContentType: []string{cc.ct.String()}}
		data, err := MarshalHeader(cc.payload, header)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(cc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = UnmarshalHeader(data, cc.newDst(), header)
			}
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
