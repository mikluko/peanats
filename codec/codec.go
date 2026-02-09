package codec

import (
	"errors"
	"fmt"
	"net/textproto"
)

// Header is a type alias for textproto.MIMEHeader, identical to peanats.Header.
// Defined here to avoid circular imports between codec and root packages.
type Header = textproto.MIMEHeader

const HeaderContentType = "Content-Type"

// Codec defines the interface for message serialization/deserialization.
type Codec interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
	ContentType() ContentType
	SetContentType(Header)
	MatchContentType(Header) bool
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

// MarshalHeader encodes x using the codec specified in the header.
func MarshalHeader(x any, header Header) ([]byte, error) {
	codec, err := ForHeader(header)
	if err != nil {
		return nil, err
	}
	if x == nil {
		return nil, nil
	}
	return codec.Marshal(x)
}

// UnmarshalHeader decodes data into x using the codec specified in the header.
func UnmarshalHeader(data []byte, x any, header Header) error {
	if data == nil {
		return nil
	}
	codec, err := ForHeader(header)
	if err != nil {
		return err
	}
	return codec.Unmarshal(data, x)
}

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
