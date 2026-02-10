package codec

import (
	"errors"
	"fmt"
)

const HeaderContentEncoding = "Content-Encoding"

// ContentEncoding represents a supported compression algorithm.
type ContentEncoding uint16

const (
	Zstd ContentEncoding = iota + 1
	S2
)

func (e ContentEncoding) String() string {
	switch e {
	case Zstd:
		return "zstd"
	case S2:
		return "s2"
	default:
		panic(fmt.Sprintf("unknown content encoding: %d", e))
	}
}

// Compressor defines the interface for message compression/decompression.
type Compressor interface {
	Compress([]byte) ([]byte, error)
	Decompress([]byte) ([]byte, error)
	ContentEncoding() ContentEncoding
}

// ErrUnsupportedEncoding is returned when a Content-Encoding value is not recognized.
var ErrUnsupportedEncoding = errors.New("unsupported content encoding")

var compressors = []Compressor{
	zstdCompressor{},
	s2Compressor{},
}

// ForContentEncoding returns the compressor for the given content encoding.
func ForContentEncoding(e ContentEncoding) (Compressor, error) {
	for _, c := range compressors {
		if c.ContentEncoding() == e {
			return c, nil
		}
	}
	return nil, fmt.Errorf("%w: %d", ErrUnsupportedEncoding, e)
}

// EncodingFromHeader extracts the ContentEncoding from a header.
// Returns 0 if absent or unrecognized.
func EncodingFromHeader(h Header) ContentEncoding {
	if h == nil {
		return 0
	}
	v := h.Get(HeaderContentEncoding)
	if v == "" {
		return 0
	}
	for _, c := range compressors {
		if v == c.ContentEncoding().String() {
			return c.ContentEncoding()
		}
	}
	return 0
}

// SetContentEncoding sets the Content-Encoding header value.
func SetContentEncoding(h Header, e ContentEncoding) {
	h.Set(HeaderContentEncoding, e.String())
}
