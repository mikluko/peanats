package codec

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

const HeaderContentType = "Content-Type"

// ContentType represents a supported serialization format.
type ContentType uint16

const (
	_ ContentType = iota
	JSON
	YAML
	Msgpack
	Protojson
	Prototext
	Protobin

	Default = JSON
)

func (c ContentType) String() string {
	switch c {
	case JSON:
		return "application/json"
	case YAML:
		return "application/yaml"
	case Msgpack:
		return "application/msgpack"
	case Protojson:
		return "application/json+proto"
	case Prototext:
		return "application/plaintext+proto"
	case Protobin:
		return "application/binary+proto"
	default:
		panic(fmt.Sprintf("unknown content type: %d", c))
	}
}

// Codec defines the interface for message serialization/deserialization.
type Codec interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
	ContentType() ContentType
	SetContentType(Header)
	MatchContentType(Header) bool
}

var codecs = []Codec{
	jsonMarshaler{},
	yamlMarshaler{},
	msgpackMarshaler{},
	protojsonCodec{
		unmarshaler: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	},
	prototextCodec{
		unmarshaler: prototext.UnmarshalOptions{
			DiscardUnknown: true,
		},
	},
	protobinCodec{
		unmarshaler: proto.UnmarshalOptions{
			DiscardUnknown: true,
		},
	},
}

// ForContentType returns the codec for the given content type.
func ForContentType(c ContentType) (Codec, error) {
	for _, codec := range codecs {
		if codec.ContentType() == c {
			return codec, nil
		}
	}
	return nil, errors.New("no codec found")
}

// ForHeader returns the codec matching the Content-Type header.
func ForHeader(h Header) (Codec, error) {
	return ForContentType(TypeFromHeader(h))
}

// TypeFromHeader extracts the ContentType from a header, defaulting to JSON.
func TypeFromHeader(header Header) ContentType {
	if header != nil {
		if v := header.Get(HeaderContentType); v != "" {
			for _, codec := range codecs {
				if v == codec.ContentType().String() {
					return codec.ContentType()
				}
			}
		}
	}
	return JSON
}

// TypeFromHeaderCopy copies the Content-Type from src to dst, defaulting to JSON.
func TypeFromHeaderCopy(dst, src Header) {
	x := src.Get(HeaderContentType)
	if x == "" {
		x = JSON.String()
	}
	dst.Set(HeaderContentType, x)
}
