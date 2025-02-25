package jetbucket

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/textproto"

	"github.com/mikluko/peanats"
)

var errUnsupportedContentType = errors.New("unsupported content type")

func chooseCodec(h textproto.MIMEHeader) (peanats.Codec, error) {
	switch h.Get("Content-Type") {
	case "application/json":
		return peanats.JsonCodec{}, nil
	case "application/protobuf":
		return peanats.ProtoCodec{}, nil
	case "application/prototext":
		return peanats.PrototextCodec{}, nil
	case "application/protojson":
		return peanats.ProtojsonCodec{}, nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedContentType, h.Get("Content-Type"))
	}
}

const (
	defaultContentType = "application/json"
	boundary           = "--"
)

func encode[T any](h textproto.MIMEHeader, v *T) ([]byte, error) {
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", defaultContentType)
	}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	if err := w.SetBoundary(boundary); err != nil {
		panic(err)
	}
	c, err := chooseCodec(h)
	if err != nil {
		return nil, err
	}
	p, err := c.Encode(v)
	if err != nil {
		return nil, err
	}
	q, _ := w.CreatePart(h)
	_, _ = q.Write(p)
	return buf.Bytes(), nil
}

func decode[T any](b []byte) (_ textproto.MIMEHeader, _ *T, err error) {
	r := multipart.NewReader(bytes.NewReader(b), boundary)
	p, err := r.NextPart()
	if err != nil {
		return nil, nil, err
	}
	h := p.Header
	c, err := chooseCodec(h)
	if err != nil {
		return nil, nil, err
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(p)
	v := new(T)
	err = c.Decode(buf.Bytes(), v)
	if err != nil {
		return nil, nil, err
	}
	return h, v, nil
}
