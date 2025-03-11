package peanats

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func Unmarshal[T any](c ContentType, p []byte, arg *T) error {
	codec, err := CodecContentType(c)
	if err != nil {
		return err
	}
	return codec.Decode(p, arg)
}

func Marshal[T any](c ContentType, arg T) ([]byte, error) {
	codec, err := CodecContentType(c)
	if err != nil {
		return nil, err
	}
	return codec.Encode(arg)
}

var codecs = []Codec{
	JsonCodec{},
	ProtojsonCodec{},
	PrototextCodec{},
	ProtobinCodec{},
}

type Codec interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, vPtr any) error
	ContentType() ContentType
	SetHeader(header Header)
	MatchHeader(header Header) bool
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

func (JsonCodec) ContentType() ContentType {
	return ContentTypeJson
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

func (ProtobinCodec) ContentType() ContentType {
	return ContentTypeProtobin
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

func (ProtojsonCodec) ContentType() ContentType {
	return ContentTypeProtojson
}

func (ProtojsonCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypeProtojson.String())
}

func (ProtojsonCodec) MatchHeader(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypeProtojson.String()
}

type ProtoyamlCodec struct{}

func (ProtoyamlCodec) Encode(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return protojson.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (ProtoyamlCodec) Decode(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return protojson.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (ProtoyamlCodec) ContentType() ContentType {
	return ContentTypeProtojson
}

func (ProtoyamlCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypeProtojson.String())
}

func (ProtoyamlCodec) MatchHeader(header Header) bool {
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

func (PrototextCodec) ContentType() ContentType {
	return ContentTypePrototext
}

func (PrototextCodec) SetHeader(header Header) {
	header.Set(HeaderContentType, ContentTypePrototext.String())
}

func (PrototextCodec) MatchHeader(header Header) bool {
	return header.Get(HeaderContentType) == ContentTypePrototext.String()
}
