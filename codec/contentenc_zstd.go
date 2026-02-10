package codec

import (
	"sync"

	"github.com/klauspost/compress/zstd"
)

var (
	zstdEnc     *zstd.Encoder
	zstdDec     *zstd.Decoder
	zstdEncOnce sync.Once
	zstdDecOnce sync.Once
)

func getZstdEncoder() *zstd.Encoder {
	zstdEncOnce.Do(func() {
		var err error
		zstdEnc, err = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
		if err != nil {
			panic(err)
		}
	})
	return zstdEnc
}

func getZstdDecoder() *zstd.Decoder {
	zstdDecOnce.Do(func() {
		var err error
		zstdDec, err = zstd.NewReader(nil)
		if err != nil {
			panic(err)
		}
	})
	return zstdDec
}

type zstdCompressor struct{}

func (zstdCompressor) Compress(data []byte) ([]byte, error) {
	return getZstdEncoder().EncodeAll(data, nil), nil
}

func (zstdCompressor) Decompress(data []byte) ([]byte, error) {
	return getZstdDecoder().DecodeAll(data, nil)
}

func (zstdCompressor) ContentEncoding() ContentEncoding {
	return Zstd
}
