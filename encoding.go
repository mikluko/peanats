package peanats

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func Unmarshal[T any](header Header, p []byte, arg *T) error {
	codec := ChooseCodec(header)
	return codec.Decode(p, arg)
}

func Marshal[T any](header Header, arg T) ([]byte, error) {
	codec := ChooseCodec(header)
	return codec.Encode(arg)
}

var codecs = []Codec{
	JsonCodec{},
	ProtojsonCodec{},
	PrototextCodec{},
	ProtobinCodec{},
}

func AddCodec(codec Codec) {
	codecs = append(codecs, codec)
}

type Codec interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, vPtr any) error
	SetHeader(header Header)
	MatchHeader(header Header) bool
}

func ChooseCodec(header Header) Codec {
	if header != nil {
		for _, codec := range codecs {
			if codec.MatchHeader(header) {
				return codec
			}
		}
	}
	return JsonCodec{}
}

type ContentType uint16

const (
	HeaderContentType = "Content-Type"

	_ ContentType = iota
	ContentTypeJson
	ContentTypeProtojson
	ContentTypePrototext
	ContentTypeProtobin
)

func (c ContentType) String() string {
	switch c {
	case ContentTypeJson:
		return "application/json"
	case ContentTypeProtojson:
		return "application/protojson"
	case ContentTypePrototext:
		return "application/prototext"
	case ContentTypeProtobin:
		return "application/protobin"
	default:
		panic(fmt.Sprintf("unknown content type: %d", c))
	}
}

type JsonCodec struct{}

func (JsonCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (JsonCodec) Decode(data []byte, vPtr any) error {
	return json.Unmarshal(data, vPtr)
}

func (JsonCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypeJson.String())
}

func (JsonCodec) MatchHeader(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeJson.String()
}

type ProtobinCodec struct{}

func (ProtobinCodec) Encode(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return proto.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (ProtobinCodec) Decode(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return proto.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (ProtobinCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypeProtobin.String())
}

func (ProtobinCodec) MatchHeader(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeProtobin.String()
}

type ProtojsonCodec struct{}

func (ProtojsonCodec) Encode(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return protojson.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (ProtojsonCodec) Decode(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return protojson.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (ProtojsonCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypeProtojson.String())
}

func (ProtojsonCodec) MatchHeader(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeProtojson.String()
}

type PrototextCodec struct{}

func (PrototextCodec) Encode(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return prototext.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (PrototextCodec) Decode(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return prototext.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (PrototextCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypePrototext.String())
}

func (PrototextCodec) MatchHeader(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypePrototext.String()
}
