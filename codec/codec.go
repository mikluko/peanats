package codec

import "net/textproto"

// Header is a type alias for textproto.MIMEHeader, identical to peanats.Header.
// Defined here to avoid circular imports between codec and root packages.
type Header = textproto.MIMEHeader

// MarshalHeader encodes x using the codec specified in the header,
// then compresses the result if Content-Encoding is set.
func MarshalHeader(x any, header Header) ([]byte, error) {
	codec, err := ForHeader(header)
	if err != nil {
		return nil, err
	}
	if x == nil {
		return nil, nil
	}
	data, err := codec.Marshal(x)
	if err != nil {
		return nil, err
	}
	if enc := EncodingFromHeader(header); enc != 0 {
		comp, err := ForContentEncoding(enc)
		if err != nil {
			return nil, err
		}
		data, err = comp.Compress(data)
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// UnmarshalHeader decompresses data if Content-Encoding is set,
// then decodes it into x using the codec specified in the header.
func UnmarshalHeader(data []byte, x any, header Header) error {
	if data == nil {
		return nil
	}
	if enc := EncodingFromHeader(header); enc != 0 {
		comp, err := ForContentEncoding(enc)
		if err != nil {
			return err
		}
		data, err = comp.Decompress(data)
		if err != nil {
			return err
		}
	}
	codec, err := ForHeader(header)
	if err != nil {
		return err
	}
	return codec.Unmarshal(data, x)
}
