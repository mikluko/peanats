package codec

import "gopkg.in/yaml.v3"

type yamlMarshaler struct{}

func (yamlMarshaler) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (yamlMarshaler) Unmarshal(data []byte, vPtr any) error {
	return yaml.Unmarshal(data, vPtr)
}

func (yamlMarshaler) ContentType() ContentType {
	return YAML
}

func (yamlMarshaler) SetContentType(header Header) {
	header.Set(HeaderContentType, YAML.String())
}

func (yamlMarshaler) MatchContentType(header Header) bool {
	return header.Get(HeaderContentType) == YAML.String()
}
