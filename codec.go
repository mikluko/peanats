package peanats

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
)

var codecs = []Codec{
	jsonMarshaler{},
	yamlMarshaler{},
	msgpackMarshaler{},
	protojsonCodec{},
	prototextCodec{},
	protobinCodec{},
}

type Codec interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
	ContentType() ContentType
	SetContentType(Header)
	MatchContentType(Header) bool
}

func CodecContentType(c ContentType) (Codec, error) {
	for _, codec := range codecs {
		if codec.ContentType() == c {
			return codec, nil
		}
	}
	return nil, errors.New("no codec found")
}

func CodecHeader(h Header) (Codec, error) {
	return CodecContentType(ContentTypeHeader(h))
}

func ContentTypeHeader(header Header) ContentType {
	if header != nil {
		if v := header.Get(HeaderContentType); v != "" {
			for _, codec := range codecs {
				if v == codec.ContentType().String() {
					return codec.ContentType()
				}
			}
		}
	}
	return ContentTypeJson
}

// MarshalHeader encodes x using the codec specified in the header.
func MarshalHeader(x any, header Header) ([]byte, error) {
	codec, err := CodecHeader(header)
	if err != nil {
		return nil, err
	}
	if header != nil {
		header.Set(HeaderContentType, codec.ContentType().String())
	}
	if x == nil {
		return nil, nil
	}
	return codec.Marshal(x)
}

func ContentTypeHeaderCopy(dst, src Header) {
	x := src.Get(HeaderContentType)
	if x == "" {
		x = ContentTypeJson.String()
	}
	dst.Set(HeaderContentType, x)
}

// UnmarshalHeader decodes data into x using the codec specified in the header.
func UnmarshalHeader(data []byte, x any, header Header) error {
	if data == nil {
		return nil
	}
	codec, err := CodecHeader(header)
	if err != nil {
		return err
	}
	return codec.Unmarshal(data, x)
}

type ContentType uint16

const (
	HeaderContentType = "Content-Type"

	_ ContentType = iota
	ContentTypeJson
	ContentTypeYaml
	ContentTypeMsgpack
	ContentTypeProtojson
	ContentTypePrototext
	ContentTypeProtobin

	DefaultContentType = ContentTypeJson
)

func (c ContentType) String() string {
	switch c {
	case ContentTypeJson:
		return "application/json"
	case ContentTypeYaml:
		return "application/yaml"
	case ContentTypeMsgpack:
		return "application/msgpack"
	case ContentTypeProtojson:
		return "application/json+proto"
	case ContentTypePrototext:
		return "application/plaintext+proto"
	case ContentTypeProtobin:
		return "application/binary+proto"
	default:
		panic(fmt.Sprintf("unknown content type: %d", c))
	}
}

type jsonMarshaler struct{}

func (jsonMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonMarshaler) Unmarshal(data []byte, vPtr any) error {
	return json.Unmarshal(data, vPtr)
}

func (jsonMarshaler) ContentType() ContentType {
	return ContentTypeJson
}

func (jsonMarshaler) SetContentType(header Header) {
	header.Set(HeaderContentType, ContentTypeJson.String())
}

func (jsonMarshaler) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeJson.String()
}

type yamlMarshaler struct{}

func (yamlMarshaler) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (yamlMarshaler) Unmarshal(data []byte, vPtr any) error {
	return yaml.Unmarshal(data, vPtr)
}

func (yamlMarshaler) ContentType() ContentType {
	return ContentTypeYaml
}

func (yamlMarshaler) SetContentType(header Header) {
	header.Set(HeaderContentType, ContentTypeYaml.String())
}

func (yamlMarshaler) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeYaml.String()
}

type msgpackMarshaler struct{}

func (msgpackMarshaler) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (msgpackMarshaler) Unmarshal(data []byte, vPtr any) error {
	return msgpack.Unmarshal(data, vPtr)
}

func (msgpackMarshaler) ContentType() ContentType {
	return ContentTypeMsgpack
}

func (msgpackMarshaler) SetContentType(header Header) {
	header.Set(HeaderContentType, ContentTypeMsgpack.String())
}

func (msgpackMarshaler) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeMsgpack.String()
}

type protobinCodec struct{}

func (protobinCodec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return proto.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (protobinCodec) Unmarshal(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return proto.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (protobinCodec) ContentType() ContentType {
	return ContentTypeProtobin
}

func (protobinCodec) SetContentType(header Header) {
	header.Set(HeaderContentType, ContentTypeProtobin.String())
}

func (protobinCodec) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeProtobin.String()
}

type protojsonCodec struct{}

func (protojsonCodec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return protojson.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (protojsonCodec) Unmarshal(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return protojson.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (protojsonCodec) ContentType() ContentType {
	return ContentTypeProtojson
}

func (protojsonCodec) SetContentType(header Header) {
	header.Set(HeaderContentType, ContentTypeProtojson.String())
}

func (protojsonCodec) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeProtojson.String()
}

type prototextCodec struct{}

func (prototextCodec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return prototext.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (prototextCodec) Unmarshal(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return prototext.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (prototextCodec) ContentType() ContentType {
	return ContentTypePrototext
}

func (prototextCodec) SetContentType(header Header) {
	header.Set(HeaderContentType, ContentTypePrototext.String())
}

func (prototextCodec) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypePrototext.String()
}
