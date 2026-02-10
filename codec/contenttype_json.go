package codec

import "encoding/json"

type jsonMarshaler struct{}

func (jsonMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonMarshaler) Unmarshal(data []byte, vPtr any) error {
	return json.Unmarshal(data, vPtr)
}

func (jsonMarshaler) ContentType() ContentType {
	return JSON
}

func (jsonMarshaler) SetContentType(header Header) {
	header.Set(HeaderContentType, JSON.String())
}

func (jsonMarshaler) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == JSON.String()
}
