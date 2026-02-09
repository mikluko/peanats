package codec

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type protobinCodec struct {
	unmarshaler proto.UnmarshalOptions
}

func (protobinCodec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return proto.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (c protobinCodec) Unmarshal(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return c.unmarshaler.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (protobinCodec) ContentType() ContentType {
	return Protobin
}

func (protobinCodec) SetContentType(header Header) {
	header.Set(HeaderContentType, Protobin.String())
}

func (protobinCodec) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == Protobin.String()
}

type protojsonCodec struct {
	unmarshaler protojson.UnmarshalOptions
}

func (protojsonCodec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return protojson.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (c protojsonCodec) Unmarshal(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return c.unmarshaler.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (protojsonCodec) ContentType() ContentType {
	return Protojson
}

func (protojsonCodec) SetContentType(header Header) {
	header.Set(HeaderContentType, Protojson.String())
}

func (protojsonCodec) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == Protojson.String()
}

type prototextCodec struct {
	unmarshaler prototext.UnmarshalOptions
}

func (prototextCodec) Marshal(v any) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return prototext.Marshal(msg)
	}
	return nil, fmt.Errorf("%T is not a proto.Message", v)
}

func (c prototextCodec) Unmarshal(data []byte, vPtr any) error {
	if msg, ok := vPtr.(proto.Message); ok {
		return c.unmarshaler.Unmarshal(data, msg)
	}
	return fmt.Errorf("%T is not a proto.Message", vPtr)
}

func (prototextCodec) ContentType() ContentType {
	return Prototext
}

func (prototextCodec) SetContentType(header Header) {
	header.Set(HeaderContentType, Prototext.String())
}

func (prototextCodec) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == Prototext.String()
}
