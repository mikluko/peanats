package codec

import "github.com/klauspost/compress/s2"

type s2Compressor struct{}

func (s2Compressor) Compress(data []byte) ([]byte, error) {
	return s2.Encode(nil, data), nil
}

func (s2Compressor) Decompress(data []byte) ([]byte, error) {
	return s2.Decode(nil, data)
}

func (s2Compressor) ContentEncoding() ContentEncoding {
	return S2
}
