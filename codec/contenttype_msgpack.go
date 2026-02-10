package codec

import "github.com/vmihailenco/msgpack/v5"

type msgpackMarshaler struct{}

func (msgpackMarshaler) Marshal(v any) ([]byte, error) {
	return msgpack.Marshal(v)
}

func (msgpackMarshaler) Unmarshal(data []byte, vPtr any) error {
	return msgpack.Unmarshal(data, vPtr)
}

func (msgpackMarshaler) ContentType() ContentType {
	return Msgpack
}

func (msgpackMarshaler) SetContentType(header Header) {
	header.Set(HeaderContentType, Msgpack.String())
}

func (msgpackMarshaler) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == Msgpack.String()
}
