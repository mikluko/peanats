package codec

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

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
